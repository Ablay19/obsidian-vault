module github.com/abdoullahelvogani/obsidian-vault/apps/api-gateway

go 1.21

require github.com/abdoullahelvogani/obsidian-vault/packages/shared-types/go v1.0.0
require github.com/abdoullahelvogani/obsidian-vault/packages/communication/go v1.0.0

replace github.com/abdoullahelvogani/obsidian-vault/packages/shared-types/go => ../../packages/shared-types/go
replace github.com/abdoullahelvogani/obsidian-vault/packages/communication/go => ../../packages/communication/go