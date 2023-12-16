package helper

import (
	"fmt"
	"myproject/model"
)

func GetEmailSubject(ticket model.Ticket) string {
	return "Pembelian Tiket Wisata Berhasil - Invoice No: " + ticket.InvoiceNumber
}

func GetEmailBody(ticket model.Ticket, totalCost int, wisataName, kodeVoucher string, pointsEarned, usedPoints int, carbonFootprint float64) string {
	emailBody := "<html><head><style>"
	emailBody += "body {font-family: Arial, sans-serif;}"
	emailBody += ".container {max-width: 600px; margin: 0 auto; padding: 20px;}"
	emailBody += ".header {background-color: #1E90FF; color: #fff; padding: 20px; border-bottom: 1px solid #ddd; position: relative;}"
	emailBody += ".header img {width: 50px; height: 50px; border-radius: 50%; border: 2px solid #fff; position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%);}"
	emailBody += "h1 {margin: 0; color: #333; font-size: 28px;}"
	emailBody += ".invoice-details {background-color: #f5f5f5; padding: 20px; margin-top: 20px; text-align: left;}"
	emailBody += "p {font-size: 16px; margin-top: 10px; color: #555; line-height: 1.5;}"
	emailBody += "strong {font-weight: bold;}"
	emailBody += "hr {border: 1px solid #ccc; margin: 20px 0;}"
	emailBody += ".footer {text-align: center; padding: 20px; color: #666; font-size: 14px; border-top: 1px solid #ddd;}"
	emailBody += "a {text-decoration: none; color: #1E90FF; font-weight: bold;}"
	emailBody += "</style></head><body>"
	emailBody += "<div class='container'>"
	emailBody += "<div class='header'>"
	emailBody += "<img src='https://i.ibb.co/KXxHL9r/Logo-Destimate.png' alt='Destimate Logo'>"
	emailBody += "<h1>Successful Ticket Purchase</h1>"
	emailBody += "</div>"
	emailBody += "<div class='invoice-details'>"
	emailBody += "<p><strong>Invoice Number:</strong> " + ticket.InvoiceNumber + "</p>"
	emailBody += "<p><strong>Destination:</strong> " + wisataName + "</p>"
	emailBody += "<p><strong>Quantity:</strong> " + fmt.Sprintf("%d", ticket.Quantity) + "</p>"
	emailBody += "<p><strong>Check-in Date:</strong> " + ticket.CheckinBooking.Format("2006-01-02") + "</p>"
	emailBody += "<p><strong>Total Price:</strong> Rp. " + fmt.Sprintf("%d", totalCost) + "</p>"

	ifStringNotEmpty := func(s, strIfNotEmpty string) string {
		if s != "" {
			return strIfNotEmpty
		}
		return ""
	}

	emailBody += ifStringNotEmpty(kodeVoucher, "<p><strong>Voucher Code:</strong> "+kodeVoucher+"</p>")
	emailBody += "<p><strong>Points Earned:</strong> " + fmt.Sprintf("%d", pointsEarned) + "</p>"
	emailBody += "<p><strong>Used Points:</strong> " + fmt.Sprintf("%d", usedPoints) + "</p>"
	emailBody += "<p><strong>Carbon Footprint:</strong> " + fmt.Sprintf("%.2f", carbonFootprint) + " grams CO2</p>"
	emailBody += "</div>"
	emailBody += "<hr>"
	emailBody += "<p>Thank you for your purchase! We hope you have a wonderful experience using our platform.</p>"
	emailBody += "<div style='text-align: left; font-size: 14px; margin-top: 20px;'>"
	emailBody += "<p style='font-weight: bold; margin: 0;'>Best Regards</p>"
	emailBody += "<p style='font-size: 12px; margin: 0;'>Muhammad Aimar Rizki Utama 💙</p>"
	emailBody += "<p style='font-size: 12px; margin: 0;'>CEO Destimate Indonesia</p>"
	emailBody += "</div>"
	emailBody += "</div>"
	emailBody += "<div class='footer'>"
	emailBody += "<p>&copy; 2023 Destimate. All rights reserved. | <a href='https://destimate-dev.netlify.app/' target='_blank'>Destimate</a></p>"
	emailBody += "</div>"
	emailBody += "</body></html>"

	return emailBody
}
