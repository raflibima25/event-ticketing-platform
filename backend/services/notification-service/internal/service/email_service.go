package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	pb "github.com/raflibima25/event-ticketing-platform/backend/pb/notification"
	"github.com/raflibima25/event-ticketing-platform/backend/services/notification-service/internal/client"
	"github.com/raflibima25/event-ticketing-platform/backend/services/notification-service/internal/template"
	"github.com/raflibima25/event-ticketing-platform/backend/services/notification-service/internal/utility"
)

// EmailService handles email sending logic
type EmailService interface {
	SendTicketEmail(ctx context.Context, req *pb.SendTicketEmailRequest) (*pb.SendTicketEmailResponse, error)
}

// emailService implements EmailService interface
type emailService struct {
	resendClient *client.ResendClient
	fromName     string
	fromEmail    string
	testMode     bool
	testEmail    string
}

// NewEmailService creates new email service instance
func NewEmailService(resendClient *client.ResendClient, fromName, fromEmail string, testMode bool, testEmail string) EmailService {
	return &emailService{
		resendClient: resendClient,
		fromName:     fromName,
		fromEmail:    fromEmail,
		testMode:     testMode,
		testEmail:    testEmail,
	}
}

// SendTicketEmail sends e-ticket email to customer with PDF attachments
func (s *emailService) SendTicketEmail(ctx context.Context, req *pb.SendTicketEmailRequest) (*pb.SendTicketEmailResponse, error) {
	log.Printf("[EmailService] Preparing ticket email for order: %s, recipient: %s, tickets: %d", req.OrderId, req.RecipientEmail, len(req.Tickets))

	// Generate PDF for each ticket
	var attachments []client.EmailAttachment
	for i, ticket := range req.Tickets {
		log.Printf("[EmailService] Generating PDF for ticket %d/%d: %s", i+1, len(req.Tickets), ticket.TicketId)

		// Prepare ticket data for PDF
		pdfData := &utility.TicketPDFData{
			TicketID:       ticket.TicketId,
			TicketNumber:   fmt.Sprintf("TKT-%s-%03d", req.OrderId[:8], i+1),
			TierName:       ticket.TierName,
			Price:          ticket.Price,
			QRCodeBase64:   ticket.QrCode,
			EventName:      req.EventName,
			EventLocation:  req.EventLocation,
			EventStartTime: req.EventStartTime,
			OrderID:        req.OrderId,
		}

		// Generate PDF
		pdfBytes, err := utility.GenerateTicketPDF(pdfData)
		if err != nil {
			log.Printf("[EmailService] Failed to generate PDF for ticket %s: %v", ticket.TicketId, err)
			return &pb.SendTicketEmailResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to generate PDF: %v", err),
			}, nil
		}

		// Encode PDF to base64 for attachment
		pdfBase64 := base64.StdEncoding.EncodeToString(pdfBytes)

		// Create attachment
		filename := fmt.Sprintf("e-ticket-%s-%s.pdf", req.OrderId[:8], ticket.TierName)
		attachments = append(attachments, client.EmailAttachment{
			Filename: filename,
			Content:  pdfBase64,
		})

		log.Printf("[EmailService] ✅ PDF generated for ticket %s (%d KB)", ticket.TicketId, len(pdfBytes)/1024)
	}

	// Build email HTML (simplified - tickets are in PDF)
	htmlContent := template.BuildTicketEmailWithPDF(&template.TicketEmailData{
		RecipientName:  req.RecipientName,
		OrderID:        req.OrderId,
		EventName:      req.EventName,
		EventLocation:  req.EventLocation,
		EventStartTime: req.EventStartTime,
		TotalAmount:    req.TotalAmount,
		PaymentMethod:  req.PaymentMethod,
		TicketCount:    len(req.Tickets),
	})

	// Determine recipient email (use test email if in test mode)
	recipientEmail := req.RecipientEmail
	if s.testMode && s.testEmail != "" {
		log.Printf("[EmailService] 🧪 Test mode enabled - redirecting email from %s to %s", req.RecipientEmail, s.testEmail)
		recipientEmail = s.testEmail
	}

	// Send email via Resend with PDF attachments
	emailReq := &client.EmailRequest{
		From:        fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail),
		To:          recipientEmail,
		Subject:     fmt.Sprintf("🎟️ E-Ticket Anda - %s", req.EventName),
		HTML:        htmlContent,
		Attachments: attachments,
	}

	emailResp, err := s.resendClient.SendEmail(emailReq)
	if err != nil {
		log.Printf("[EmailService] Failed to send email for order %s: %v", req.OrderId, err)
		return &pb.SendTicketEmailResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to send email: %v", err),
		}, nil
	}

	log.Printf("[EmailService] ✅ Email sent successfully for order %s with %d PDF attachments, email ID: %s", req.OrderId, len(attachments), emailResp.ID)

	return &pb.SendTicketEmailResponse{
		Success: true,
		Message: "E-ticket email sent successfully with PDF attachments",
		EmailId: emailResp.ID,
	}, nil
}
