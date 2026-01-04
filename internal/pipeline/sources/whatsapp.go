package sources

import (
	"context"
	"fmt"
	"log/slog"
	"obsidian-automation/internal/dashboard/ws"
	"obsidian-automation/internal/pipeline"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
)

type WhatsAppSource struct {
	client    *whatsmeow.Client
	wsManager *ws.Manager
	dbLog     *slog.Logger
}

func NewWhatsAppSource(wsManager *ws.Manager) *WhatsAppSource {
	return &WhatsAppSource{
		wsManager: wsManager,
		dbLog:     slog.With("source", "whatsapp"),
	}
}

func (s *WhatsAppSource) Name() string {
	return "whatsapp"
}

func (s *WhatsAppSource) Start(ctx context.Context, jobChan chan<- pipeline.Job) error {
	dbLog := slog.With("source", "whatsapp")
	dbLog.Info("Initializing WhatsApp client")

	container, err := sqlstore.New(ctx, "sqlite3", "file:whatsapp_session.db?_foreign_keys=on", nil)
	if err != nil {
		return fmt.Errorf("failed to connect to session db: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return fmt.Errorf("failed to get device store: %w", err)
	}

	client := whatsmeow.NewClient(deviceStore, nil)
	s.client = client

	// Handle Messages
	client.AddEventHandler(func(evt interface{}) {
		s.handleEvent(ctx, evt, jobChan)
	})

	// Handle Connection Lifecycle
	client.AddEventHandler(func(evt interface{}) {
		switch evt.(type) {
		case *events.Connected:
			dbLog.Info("WhatsApp connected")
			if s.wsManager != nil {
				s.wsManager.Broadcast("whatsapp_connected", map[string]bool{"connected": true})
			}
		case *events.LoggedOut:
			dbLog.Info("WhatsApp logged out")
			if s.wsManager != nil {
				s.wsManager.Broadcast("whatsapp_connected", map[string]bool{"connected": false})
			}
		}
	})

	if client.Store.ID == nil {
		// No ID stored, need to link device
		if s.wsManager != nil {
			s.wsManager.Broadcast("whatsapp_connecting", map[string]bool{"linking": true})
		}
		qrChan, _ := client.GetQRChannel(ctx)
		err = client.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}

		go func() {
			for {
				select {
				case evt, ok := <-qrChan:
					if !ok {
						return
					}
					if evt.Event == "code" {
						dbLog.Info("QR code received", "code", evt.Code)
						if s.wsManager != nil {
							s.wsManager.Broadcast("whatsapp_qr", map[string]string{
								"code": evt.Code,
							})
						}
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	} else {
		// Already logged in, just connect
		if s.wsManager != nil {
			s.wsManager.Broadcast("whatsapp_connecting", map[string]bool{"connecting": true})
		}
		err = client.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
	}

	// Listen for stop signal to disconnect gracefully
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-c:
		case <-ctx.Done():
		}
		client.Disconnect()
	}()

	return nil
}

func (s *WhatsAppSource) handleEvent(ctx context.Context, evt interface{}, jobChan chan<- pipeline.Job) {
	switch v := evt.(type) {
	case *events.Message:
		s.handleMessage(ctx, v, jobChan)
	}
}

func (s *WhatsAppSource) handleMessage(ctx context.Context, evt *events.Message, jobChan chan<- pipeline.Job) {
	if evt.Info.IsFromMe {
		return
	}

	text := ""
	var data []byte
	var contentType pipeline.ContentType = pipeline.ContentTypeText

	// Handle different message types
	if msg := evt.Message.GetConversation(); msg != "" {
		text = msg
	} else if msg := evt.Message.GetExtendedTextMessage().GetText(); msg != "" {
		text = msg
	} else if img := evt.Message.GetImageMessage(); img != nil {
		s.dbLog.Info("Received image from WhatsApp", "sender", evt.Info.Sender)
		downloaded, err := s.client.Download(ctx, img)
		if err == nil {
			data = downloaded
			contentType = pipeline.ContentTypeImage
			text = img.GetCaption()
		} else {
			s.dbLog.Error("Failed to download WhatsApp image", "error", err)
		}
	} else if doc := evt.Message.GetDocumentMessage(); doc != nil {
		if doc.GetMimetype() == "application/pdf" {
			s.dbLog.Info("Received PDF from WhatsApp", "sender", evt.Info.Sender)
			downloaded, err := s.client.Download(ctx, doc)
			if err == nil {
				data = downloaded
				contentType = pipeline.ContentTypePDF
				text = doc.GetCaption()
			} else {
				s.dbLog.Error("Failed to download WhatsApp PDF", "error", err)
			}
		}
	}

	if text == "" && data == nil {
		return
	}

	job := pipeline.Job{
		ID:          fmt.Sprintf("wa_%s", evt.Info.ID),
		Source:      "whatsapp",
		SourceID:    evt.Info.ID,
		Data:        data,
		ContentType: contentType,
		ReceivedAt:  time.Now(),
		MaxRetries:  3,
		UserContext: pipeline.UserContext{
			UserID:   evt.Info.Sender.String(),
			Language: "English", // Default
		},
		Metadata: map[string]interface{}{
			"caption": text,
			"sender":  evt.Info.Sender.String(),
		},
	}

	// Submit to pipeline
	select {
	case jobChan <- job:
		s.dbLog.Info("Job submitted from WhatsApp", "job_id", job.ID)
	default:
		s.dbLog.Warn("Pipeline job channel full, dropping WhatsApp message")
	}
}
