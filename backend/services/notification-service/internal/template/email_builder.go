package template

import (
	"fmt"
)

// TicketEmailData represents data for ticket email template
type TicketEmailData struct {
	RecipientName  string
	OrderID        string
	EventName      string
	EventLocation  string
	EventStartTime string
	TotalAmount    float64
	PaymentMethod  string
	Tickets        []TicketData
	TicketCount    int
}

// TicketData represents individual ticket data
type TicketData struct {
	TicketID   string
	TierName   string
	Price      float64
	QRCodeBase64 string
}

// BuildTicketEmail builds HTML email for e-tickets
func BuildTicketEmail(data *TicketEmailData) string {
	ticketsHTML := ""
	for _, ticket := range data.Tickets {
		ticketsHTML += buildTicketCard(ticket)
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>E-Ticket Anda</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background-color: #f4f4f4;
            margin: 0;
            padding: 20px;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background-color: #ffffff;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            padding: 30px 20px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 28px;
        }
        .content {
            padding: 30px 20px;
        }
        .greeting {
            font-size: 18px;
            color: #333;
            margin-bottom: 20px;
        }
        .event-info {
            background-color: #f8f9fa;
            border-left: 4px solid #667eea;
            padding: 20px;
            margin: 20px 0;
        }
        .event-info h2 {
            margin: 0 0 15px 0;
            color: #667eea;
            font-size: 22px;
        }
        .event-detail {
            margin: 10px 0;
            color: #555;
        }
        .event-detail strong {
            color: #333;
        }
        .ticket-card {
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            padding: 20px;
            margin: 20px 0;
            background-color: #fff;
        }
        .ticket-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
            padding-bottom: 15px;
            border-bottom: 2px dashed #e0e0e0;
        }
        .ticket-tier {
            font-size: 18px;
            font-weight: bold;
            color: #667eea;
        }
        .ticket-price {
            font-size: 16px;
            color: #666;
        }
        .qr-code-container {
            text-align: center;
            padding: 20px 0;
        }
        .qr-code-container img {
            max-width: 200px;
            height: auto;
        }
        .ticket-id {
            text-align: center;
            font-size: 12px;
            color: #999;
            font-family: 'Courier New', monospace;
            margin-top: 10px;
        }
        .order-summary {
            background-color: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            margin: 20px 0;
        }
        .summary-row {
            display: flex;
            justify-content: space-between;
            margin: 10px 0;
        }
        .summary-row.total {
            font-weight: bold;
            font-size: 18px;
            color: #667eea;
            border-top: 2px solid #e0e0e0;
            padding-top: 15px;
            margin-top: 15px;
        }
        .instructions {
            background-color: #fff3cd;
            border-left: 4px solid #ffc107;
            padding: 15px;
            margin: 20px 0;
        }
        .instructions h3 {
            margin: 0 0 10px 0;
            color: #856404;
        }
        .instructions ul {
            margin: 10px 0;
            padding-left: 20px;
        }
        .instructions li {
            margin: 5px 0;
            color: #856404;
        }
        .footer {
            background-color: #f8f9fa;
            padding: 20px;
            text-align: center;
            color: #666;
            font-size: 14px;
        }
        @media only screen and (max-width: 600px) {
            .ticket-header {
                flex-direction: column;
                align-items: flex-start;
            }
            .ticket-price {
                margin-top: 5px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎟️ E-Ticket Anda</h1>
        </div>

        <div class="content">
            <div class="greeting">
                Halo <strong>%s</strong>! 👋
            </div>

            <p>Terima kasih atas pembelian tiket Anda. Pembayaran telah berhasil dikonfirmasi!</p>

            <div class="event-info">
                <h2>📅 Detail Event</h2>
                <div class="event-detail">
                    <strong>Nama Event:</strong> %s
                </div>
                <div class="event-detail">
                    <strong>Lokasi:</strong> %s
                </div>
                <div class="event-detail">
                    <strong>Waktu:</strong> %s
                </div>
            </div>

            <h3 style="margin-top: 30px; color: #333;">🎫 Tiket Anda</h3>
            %s

            <div class="order-summary">
                <div class="summary-row">
                    <span>Order ID:</span>
                    <span style="font-family: 'Courier New', monospace;">%s</span>
                </div>
                <div class="summary-row">
                    <span>Metode Pembayaran:</span>
                    <span>%s</span>
                </div>
                <div class="summary-row total">
                    <span>Total Pembayaran:</span>
                    <span>Rp %s</span>
                </div>
            </div>

            <div class="instructions">
                <h3>📋 Instruksi Penting</h3>
                <ul>
                    <li>Tunjukkan <strong>QR Code</strong> di atas kepada petugas di pintu masuk</li>
                    <li>Pastikan QR Code terlihat jelas (screenshot atau print)</li>
                    <li>Datang <strong>minimal 30 menit</strong> sebelum acara dimulai</li>
                    <li>Satu tiket hanya berlaku untuk <strong>satu kali masuk</strong></li>
                    <li>Simpan email ini sebagai bukti pembelian</li>
                </ul>
            </div>

            <p style="color: #666; font-size: 14px; margin-top: 20px;">
                Jika ada pertanyaan, silakan hubungi customer service kami.
            </p>
        </div>

        <div class="footer">
            <p>Event Ticketing Platform</p>
            <p style="font-size: 12px; margin-top: 10px;">
                Email ini dikirim secara otomatis, mohon tidak membalas email ini.
            </p>
        </div>
    </div>
</body>
</html>
	`,
		data.RecipientName,
		data.EventName,
		data.EventLocation,
		data.EventStartTime,
		ticketsHTML,
		data.OrderID,
		data.PaymentMethod,
		formatCurrency(data.TotalAmount),
	)
}

func buildTicketCard(ticket TicketData) string {
	return fmt.Sprintf(`
            <div class="ticket-card">
                <div class="ticket-header">
                    <div class="ticket-tier">%s</div>
                    <div class="ticket-price">Rp %s</div>
                </div>
                <div class="qr-code-container">
                    <img src="%s" alt="QR Code">
                </div>
                <div class="ticket-id">ID: %s</div>
            </div>
	`,
		ticket.TierName,
		formatCurrency(ticket.Price),
		ticket.QRCodeBase64,
		ticket.TicketID,
	)
}

// BuildTicketEmailWithPDF builds HTML email for e-tickets with PDF attachments
func BuildTicketEmailWithPDF(data *TicketEmailData) string {
	ticketWord := "tiket"
	if data.TicketCount > 1 {
		ticketWord = "tiket"
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>E-Ticket Anda</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background-color: #f4f4f4;
            margin: 0;
            padding: 20px;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background-color: #ffffff;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            padding: 30px 20px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 28px;
        }
        .content {
            padding: 30px 20px;
        }
        .greeting {
            font-size: 18px;
            color: #333;
            margin-bottom: 20px;
        }
        .event-info {
            background-color: #f8f9fa;
            border-left: 4px solid #667eea;
            padding: 20px;
            margin: 20px 0;
        }
        .event-info h2 {
            margin: 0 0 15px 0;
            color: #667eea;
            font-size: 22px;
        }
        .event-detail {
            margin: 10px 0;
            color: #555;
        }
        .event-detail strong {
            color: #333;
        }
        .pdf-notice {
            background-color: #d1ecf1;
            border-left: 4px solid #0c5460;
            padding: 20px;
            margin: 20px 0;
            border-radius: 4px;
        }
        .pdf-notice h3 {
            margin: 0 0 10px 0;
            color: #0c5460;
            font-size: 18px;
        }
        .pdf-notice p {
            margin: 5px 0;
            color: #0c5460;
        }
        .pdf-icon {
            font-size: 48px;
            text-align: center;
            margin: 10px 0;
        }
        .order-summary {
            background-color: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            margin: 20px 0;
        }
        .summary-row {
            display: flex;
            justify-content: space-between;
            margin: 10px 0;
        }
        .summary-row.total {
            font-weight: bold;
            font-size: 18px;
            color: #667eea;
            border-top: 2px solid #e0e0e0;
            padding-top: 15px;
            margin-top: 15px;
        }
        .instructions {
            background-color: #fff3cd;
            border-left: 4px solid #ffc107;
            padding: 15px;
            margin: 20px 0;
        }
        .instructions h3 {
            margin: 0 0 10px 0;
            color: #856404;
        }
        .instructions ul {
            margin: 10px 0;
            padding-left: 20px;
        }
        .instructions li {
            margin: 5px 0;
            color: #856404;
        }
        .footer {
            background-color: #f8f9fa;
            padding: 20px;
            text-align: center;
            color: #666;
            font-size: 14px;
        }
        .download-button {
            display: inline-block;
            background-color: #667eea;
            color: white;
            padding: 12px 30px;
            text-decoration: none;
            border-radius: 5px;
            font-weight: bold;
            margin: 10px 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎟️ E-Ticket Anda</h1>
        </div>

        <div class="content">
            <div class="greeting">
                Halo <strong>%s</strong>! 👋
            </div>

            <p>Terima kasih atas pembelian tiket Anda. Pembayaran telah berhasil dikonfirmasi!</p>

            <div class="event-info">
                <h2>📅 Detail Event</h2>
                <div class="event-detail">
                    <strong>Nama Event:</strong> %s
                </div>
                <div class="event-detail">
                    <strong>Lokasi:</strong> %s
                </div>
                <div class="event-detail">
                    <strong>Waktu:</strong> %s
                </div>
            </div>

            <div class="pdf-notice">
                <h3>📎 E-Ticket Anda</h3>
                <div class="pdf-icon">📄</div>
                <p><strong>%d %s Anda terlampir dalam file PDF</strong></p>
                <p>Silakan buka file PDF yang terlampir di email ini untuk melihat e-ticket Anda lengkap dengan QR code.</p>
                <p style="margin-top: 15px; font-size: 14px;">
                    💡 <strong>Tip:</strong> Simpan file PDF ke smartphone Anda atau print untuk memudahkan saat masuk event.
                </p>
            </div>

            <div class="order-summary">
                <div class="summary-row">
                    <span>Order ID:</span>
                    <span style="font-family: 'Courier New', monospace;">%s</span>
                </div>
                <div class="summary-row">
                    <span>Jumlah Tiket:</span>
                    <span>%d %s</span>
                </div>
                <div class="summary-row">
                    <span>Metode Pembayaran:</span>
                    <span>%s</span>
                </div>
                <div class="summary-row total">
                    <span>Total Pembayaran:</span>
                    <span>Rp %s</span>
                </div>
            </div>

            <div class="instructions">
                <h3>📋 Instruksi Penting</h3>
                <ul>
                    <li>Buka file PDF e-ticket yang terlampir</li>
                    <li>Tunjukkan <strong>QR Code di PDF</strong> kepada petugas di pintu masuk</li>
                    <li>Pastikan QR Code terlihat jelas (screenshot atau print)</li>
                    <li>Datang <strong>minimal 30 menit</strong> sebelum acara dimulai</li>
                    <li>Satu tiket hanya berlaku untuk <strong>satu kali masuk</strong></li>
                    <li>Simpan email dan PDF ini sebagai bukti pembelian</li>
                </ul>
            </div>

            <p style="color: #666; font-size: 14px; margin-top: 20px;">
                Jika ada pertanyaan, silakan hubungi customer service kami.
            </p>
        </div>

        <div class="footer">
            <p>Event Ticketing Platform</p>
            <p style="font-size: 12px; margin-top: 10px;">
                Email ini dikirim secara otomatis, mohon tidak membalas email ini.
            </p>
        </div>
    </div>
</body>
</html>
	`,
		data.RecipientName,
		data.EventName,
		data.EventLocation,
		data.EventStartTime,
		data.TicketCount,
		ticketWord,
		data.OrderID,
		data.TicketCount,
		ticketWord,
		data.PaymentMethod,
		formatCurrency(data.TotalAmount),
	)
}

func formatCurrency(amount float64) string {
	// Simple currency formatting for Indonesian Rupiah
	str := fmt.Sprintf("%.0f", amount)

	// Add thousand separators
	var result []rune
	count := 0

	for i := len(str) - 1; i >= 0; i-- {
		if count > 0 && count%3 == 0 {
			result = append([]rune{'.'}, result...)
		}
		result = append([]rune{rune(str[i])}, result...)
		count++
	}

	return string(result)
}
