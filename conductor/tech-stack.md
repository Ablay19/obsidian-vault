# Technology Stack: AI-Powered Obsidian Automation Bot

## 1. Core Programming Language

*   **Go (Golang):** Chosen for its efficiency, strong concurrency model, and suitability for building high-performance backend services and APIs.

## 2. Artificial Intelligence (AI) Providers

The bot is designed to be flexible and leverage multiple AI providers for diverse capabilities:

*   **Google Gemini:** Utilized for advanced content summarization, question answering, and structured data generation, offering robust natural language processing capabilities.
*   **Groq:** Integrated for high-speed inference, particularly beneficial for real-time streaming responses and minimizing latency in AI interactions.
*   **Hugging Face:** Included for broad access to a variety of pre-trained models, allowing for potential future expansion into specialized NLP tasks or local AI processing.
*   **OpenRouter:** Integrated to provide a unified API for accessing a wide array of large language models from different providers, ensuring flexibility and redundancy.

## 3. Database

*   **Turso:** Employed as the primary database solution for persistent state management, including chat history, bot instance data, and configuration settings. Its lightweight and embedded-first approach makes it suitable for containerized deployments.

## 4. Deployment and Orchestration

*   **Docker:** The entire application is containerized using Docker, ensuring portability, consistent environments across development and production, and simplified deployment.

## 5. Text Extraction and Document Processing

*   **Tesseract OCR:** Used for optical character recognition to extract text from image files, enabling the bot to process visual content.
*   **Poppler (pdftotext):** Leveraged for efficient and accurate text extraction from PDF documents, ensuring that content from various document types can be analyzed by the AI.

## 6. Web Dashboard Technologies

*   **Go (net/http):** Powers the backend of the web dashboard.
*   **Templ:** Used for server-side HTML templating, ensuring a type-safe and efficient way to generate the dashboard's user interface.
*   **HTMX:** Enhances the dashboard's interactivity by allowing server-rendered HTML fragments to be swapped in the browser, providing a dynamic user experience without extensive JavaScript.
