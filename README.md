# TradeBook LM API

A Go-based API for trade analysis and AI-powered insights using Google Cloud Platform services.

## Features

- **User Authentication**: Firebase Auth integration
- **Trade Management**: Create, read, update, delete trades and tradebooks
- **AI Analysis**: Gemini 1.5 Flash integration for trade analysis
- **Subscription Tiers**: Rate limiting based on subscription levels
- **Payment Processing**: Stripe integration for subscriptions

## AI Rate Limiting

The API implements intelligent rate limiting for AI requests based on subscription tiers:

### Hobby Users (Free)
- **5 requests per day**
- **2 requests per hour**
- **50 requests per month**

### Basic Users
- **20 requests per day**
- **5 requests per hour**
- **200 requests per month**

### Pro Users
- **100 requests per day**
- **20 requests per hour**
- **1000 requests per month**

### Ultra Users
- **Unlimited requests**

## API Endpoints

### AI Endpoints
- `POST /vertex-ai/:model` - Generate AI analysis with rate limiting (supports gemini-1.5-flash, gemini-2.5-flash, gemini-2.5-pro-max)
- `GET /ai/usage` - Get current AI usage and limits

### Authentication
- `POST /auth` - Sign in
- `DELETE /auth` - Sign out
- `GET /auth` - Verify authentication

### Tradebooks
- `POST /tradebook` - Create tradebook
- `GET /tradebook` - List tradebooks
- `GET /tradebook/:id` - Get specific tradebook
- `PATCH /tradebook/:id` - Update tradebook
- `DELETE /tradebook/:id` - Delete tradebook

### Trades
- `POST /trade/:tradebookId` - Add trades to tradebook
- `GET /trade/:tradebookId` - Get trades from tradebook
- `PATCH /trade/:tradebookId/:tradeId` - Update specific trade
- `DELETE /trade/:tradebookId` - Delete all trades from tradebook

### Payments
- `POST /stripe-create-checkout-session` - Create Stripe checkout session

## Environment Variables

### Required
- `FIREBASE_SERVICE_ACCOUNT_SECRET` - Firebase service account JSON
- `VERTEX_AI_SERVICE_ACCOUNT_SECRET` - Vertex AI service account JSON
- `GCP_PROJECT_ID` - Google Cloud Project ID
- `GCP_REGION` - Google Cloud Region
- `STRIPE_SECRET_KEY` - Stripe secret key
- `STRIPE_MONTHLY_PRICE_ID` - Stripe monthly price ID
- `STRIPE_YEARLY_PRICE_ID` - Stripe yearly price ID
- `WEB_URL` - Frontend URL for redirects

### Optional (for local development)
- `FIREBASE_SERVICE_ACCOUNT_KEY` - Path to Firebase service account file
- `VERTEX_AI_SERVICE_ACCOUNT_KEY` - Path to Vertex AI service account file

## Rate Limiting Logic

The AI rate limiting system:

1. **Validates limits** before processing each AI request
2. **Tracks usage** in Firestore with automatic counter resets
3. **Provides clear error messages** when limits are exceeded
4. **Supports unlimited access** for premium tiers
5. **Resets counters** automatically at appropriate intervals:
   - Daily: Midnight each day
   - Hourly: Every hour from first request
   - Monthly: First day of each month

## Usage Example

```bash
# Get current AI usage
curl -H "Authorization: Bearer <token>" \
     http://localhost:8080/ai/usage

# Make AI request (with rate limiting)
curl -X POST -H "Authorization: Bearer <token>" \
     -H "Content-Type: application/json" \
     -d '{"prompt": "Analyze my trading performance", "trades": {...}}' \
     http://localhost:8080/vertex-ai/gemini-1.5-flash
```

## Development

```bash
# Run locally
go run cmd/app/main.go

# Build for production
go build -o tradebooklm-api cmd/app/main.go
```