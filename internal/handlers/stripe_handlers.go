package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
)

func StripeCreateCheckoutSessionHandler(c *gin.Context, stripeKey string) {
	stripe.Key = stripeKey

	type PlanRequest struct {
		Plan        string `json:"plan"`
		RedirectUrl string `json:"redirectUrl"`
	}

	var req PlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var priceID string

	switch req.Plan {
	case "pro_monthly":
		priceID = os.Getenv("STRIPE_PRO_MONTHLY_PRICE_ID")
	case "ultra_monthly":
		priceID = os.Getenv("STRIPE_ULTRA_MONTHLY_PRICE_ID")
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		// SuccessURL:        stripe.String(fmt.Sprintf("%s/payment-success?session_id={CHECKOUT_SESSION_ID}", os.Getenv("WEB_URL"))),
		// CancelURL:         stripe.String(fmt.Sprintf("%s/account?status=cancelled", os.Getenv("WEB_URL"))),
		// use redirectUrl from request
		SuccessURL: stripe.String(fmt.Sprintf("%s?session_id={CHECKOUT_SESSION_ID}", req.RedirectUrl)),
		CancelURL:  stripe.String(fmt.Sprintf("%s?status=cancelled", req.RedirectUrl)),
		// ClientReferenceID: stripe.String(uid),
		// CustomerEmail:     stripe.String(user.Email),
	}

	sess, err := session.New(params)
	if err != nil {
		log.Printf("Error creating checkout session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checkout session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessionId": sess.ID})
}

func StripeWebHookHandler(c *gin.Context, stripeKey string) {
	stripe.Key = stripeKey

	type PlanRequest struct {
		Plan        string `json:"plan"`
		RedirectUrl string `json:"redirectUrl"`
	}

	var req PlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var priceID string

	switch req.Plan {
	case "pro_monthly":
		priceID = os.Getenv("STRIPE_PRO_MONTHLY_PRICE_ID")
	case "ultra_monthly":
		priceID = os.Getenv("STRIPE_ULTRA_MONTHLY_PRICE_ID")
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		// SuccessURL:        stripe.String(fmt.Sprintf("%s/payment-success?session_id={CHECKOUT_SESSION_ID}", os.Getenv("WEB_URL"))),
		// CancelURL:         stripe.String(fmt.Sprintf("%s/account?status=cancelled", os.Getenv("WEB_URL"))),
		// use redirectUrl from request
		SuccessURL: stripe.String(fmt.Sprintf("%s?session_id={CHECKOUT_SESSION_ID}", req.RedirectUrl)),
		CancelURL:  stripe.String(fmt.Sprintf("%s?status=cancelled", req.RedirectUrl)),
		// ClientReferenceID: stripe.String(uid),
		// CustomerEmail:     stripe.String(user.Email),
	}

	sess, err := session.New(params)
	if err != nil {
		log.Printf("Error creating checkout session: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checkout session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessionId": sess.ID})
}
