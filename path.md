# Project Structure

```
AgroLink_API/
├── collection
│   └── Agro Link API.postman_collection.json
├── config
│   ├── config.go
│   ├── database.go
│   ├── jwt.go
│   ├── midtrans.go
│   └── migration.go
├── dto
│   ├── admin_dto.go
│   ├── application_response.go
│   ├── auth_response.go
│   ├── cart_dto.go
│   ├── contract_response.go
│   ├── dashboard_dto.go
│   ├── delivery_dto.go
│   ├── driver_response.go
│   ├── gemini_chat_dto.go
│   ├── offer_dto.go
│   ├── pagination_dto.go
│   ├── payout_dto.go
│   ├── product_dto.go
│   ├── profit_dto.go
│   ├── project_response.go
│   ├── review_dto.go
│   ├── tracking_response.go
│   ├── transaction_response.go
│   ├── user_response.go
│   └── worker_response.go
├── handlers
│   ├── admin_handlers.go
│   ├── application_handlers.go
│   ├── auth_handlers.go
│   ├── cart_handlers.go
│   ├── chat_handlers.go
│   ├── checkout_handlers.go
│   ├── contract_handlers.go
│   ├── delivery_handlers.go
│   ├── driver_handlers.go
│   ├── ecommerce_webhook_handlers.go
│   ├── farm_handlers.go
│   ├── gemini_chat_handlers.go
│   ├── notification_handlers.go
│   ├── offer_handlers.go
│   ├── payment_handlers.go
│   ├── product_handlers.go
│   ├── profile_handlers.go
│   ├── profit_handler.go
│   ├── project_handlers.go
│   ├── review_handlers.go
│   ├── tracking_handlers.go
│   ├── webhook_handlers.go
│   └── worker_handlers.go
├── middleware
│   ├── auth.go
│   └── role.go
├── models
│   ├── ai_chat.go
│   ├── cart.go
│   ├── contract.go
│   ├── conversation.go
│   ├── delivery.go
│   ├── driver_route.go
│   ├── invoice.go
│   ├── location_track.go
│   ├── message.go
│   ├── models.go
│   ├── notification.go
│   ├── order_item.go
│   ├── order.go
│   ├── participant.go
│   ├── payment_ecommerce.go
│   ├── payout.go
│   ├── platform_profit.go
│   ├── product.go
│   ├── project.go
│   ├── review.go
│   ├── schedule.go
│   ├── system.go
│   ├── transaction.go
│   ├── user_verifications.go
│   ├── user.go
│   └── webhook_log.go
├── pkg
│   └── chat
│       ├── client.go
│       ├── hub.go
│       └── message.go
├── public
│   └── images
│       └── profiles
│           ├── 73034e92-c7b3-47b2-b822-0e25eaaa7f75-78dbdf95-8ca5-4ed3-bae5-df1dad17faf7.png
│           ├── 990b2abd-ae13-4445-ba7a-4fbed6790eb2-fdf2b4e6-afa4-4973-82cf-708d3770c20c.png
│           └── c72dad24-5dac-4588-a787-88229195d59c-6b704927-ecae-48a7-97a2-7f3510fce9bc.png
├── repositories
│   ├── application_repositories.go
│   ├── assignment_repositories.go
│   ├── cart_repositories.go
│   ├── contract_repositories.go
│   ├── delivery_repositories.go
│   ├── delivery_route_repositories.go
│   ├── driver_repositories.go
│   ├── farm_repositories.go
│   ├── gemini_chat_repositories.go
│   ├── invoice_repositories.go
│   ├── location_track_repositories.go
│   ├── notification_repositories.go
│   ├── order_repositories.go
│   ├── payment_ecommerce_repositories.go
│   ├── payout_repositories.go
│   ├── product_repositories.go
│   ├── profit_repositories.go
│   ├── project_repositories.go
│   ├── review_repositories.go
│   ├── transaction_repositories.go
│   ├── user_repositories.go
│   ├── user_verification_repositories.go
│   ├── webhook_repositories.go
│   └── worker_repositories.go
├── routes
│   ├── protected_routes.go
│   └── public_routes.go
├── seeders
│   ├── seeder_users.go
│   ├── seedInProgressDeliveryScenario.go
│   ├── transaction_seeder.go
│   ├── transactions_seed_ecommerce.json
│   ├── transactions_seed_utama.json
│   └── users_seed.json
├── services
│   ├── admin_service.go
│   ├── application_service.go
│   ├── auth_service.go
│   ├── cart_service.go
│   ├── checkout_service.go
│   ├── contract_service.go
│   ├── delivery_service.go
│   ├── driver_service.go
│   ├── ecommerce_payment_service.go
│   ├── email_service.go
│   ├── farm_service.go
│   ├── gemini_chat_service.go
│   ├── notification_service.go
│   ├── offer_service.go
│   ├── payment_service.go
│   ├── product_service.go
│   ├── profile_service.go
│   ├── profit_service.go
│   ├── project_service.go
│   ├── review_service.go
│   ├── tracking_service.go
│   ├── user_service.go
│   └── worker_service.go
├── templates
│   └── contract_template.html
├── utils
│   ├── encryption.go
│   ├── response.go
│   └── send_email.go
├── AgroLink_End_to_End_Test.postman_collection.json
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── main.go
└── PROJECT_CONTEXT.md
```
