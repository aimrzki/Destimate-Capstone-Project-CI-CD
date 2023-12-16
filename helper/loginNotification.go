package helper

import (
	"github.com/go-gomail/gomail"
	"os"
	"strconv"
)

func SendLoginNotification(userEmail string, name string) error {
	// Mengambil nilai dari environment variables
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	// Konfigurasi pengiriman email
	sender := smtpUsername
	recipient := userEmail
	subject := "Successful Login Notification"
	emailBody := `
	<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login Notification</title>
    <style>
        body {
            font-family: 'Arial', sans-serif;
            background: linear-gradient(180deg, #007BFF, #00BFFF);
            color: #fff;
            margin: 0;
            padding: 0;
            display: flex;
            align-items: center;
            justify-content: center;
            height: 100vh;
        }
        .container {
            max-width: 600px;
            width: 100%;
            background-color: #fff;
            box-shadow: 0 0 20px rgba(0, 0, 0, 0.2);
            border-radius: 10px;
            overflow: hidden;
            text-align: center;
            margin: 0 auto; /* Menempatkan container di tengah */
        }
        .header {
            background-color: #007BFF;
            color: #fff;
            padding: 20px;
            border-bottom: 1px solid #ddd;
        }
        h1 {
            margin: 0;
            color: #333;
            font-size: 28px;
        }
        .logo {
            text-align: center;
            margin-top: 20px;
        }
        .logo img {
            width: 120px;
            height: 120px;
            border-radius: 50%;
            border: 3px solid #007BFF;
            transition: transform 0.3s ease-in-out;
        }
        .logo img:hover {
            transform: scale(1.1);
        }
        .message {
            padding: 20px;
        }
        p {
            font-size: 18px;
            margin-top: 15px;
            color: #555;
            line-height: 1.5;
        }
        .footer {
            text-align: center;
            padding: 20px;
            color: #666;
            font-size: 14px;
            border-top: 1px solid #ddd;
        }
        a {
            text-decoration: none;
            color: #007BFF;
            font-weight: bold;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Login Successful</h1>
        </div>
        <div class="logo">
            <img src="https://i.ibb.co/KXxHL9r/Logo-Destimate.png" alt="Destimate Logo">
        </div>
        <div class="message">
            <p>Hello, <strong>` + name + `</strong>,</p>
            <p>Your login was successful. If this wasn't you, please contact our support team immediately. Thank you.</p>
            <p><strong>Support Team:</strong> <a href="mailto:hidestimate@gmail.com">hidestimate@gmail.com@gmail.com</a></p>
        </div>
        <div class="footer">
            <p>&copy; 2023 Destimate. All rights reserved. | <a href="https://destimate-dev.netlify.app/" target="_blank">Destimate</a></p>
        </div>
    </div>
</body>
</html>




	`

	// Convert the SMTP port from string to integer
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)
	// Set pesan dalam format HTML
	m.SetBody("text/html", emailBody)

	d := gomail.NewDialer(smtpServer, smtpPort, smtpUsername, smtpPassword)

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
