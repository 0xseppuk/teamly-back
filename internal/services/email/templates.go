package email

import "fmt"

// GetPasswordResetEmailHTML возвращает HTML шаблон письма для сброса пароля
func GetPasswordResetEmailHTML(resetURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Сброс пароля - Teamly</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #0a0a0a;">
    <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0" style="background-color: #0a0a0a;">
        <tr>
            <td align="center" style="padding: 40px 20px;">
                <table role="presentation" width="600" cellspacing="0" cellpadding="0" border="0" style="background: linear-gradient(180deg, rgba(139, 92, 246, 0.1) 0%%, rgba(0, 0, 0, 0) 100%%); background-color: #18181b; border-radius: 16px; border: 1px solid rgba(255, 255, 255, 0.1);">
                    <!-- Header -->
                    <tr>
                        <td align="center" style="padding: 40px 40px 20px;">
                            <h1 style="margin: 0; font-size: 32px; font-weight: bold; color: #ffffff; letter-spacing: -0.5px;">Teamly</h1>
                        </td>
                    </tr>

                    <!-- Content -->
                    <tr>
                        <td style="padding: 20px 40px;">
                            <h2 style="margin: 0 0 16px; font-size: 24px; font-weight: 600; color: #ffffff;">Сброс пароля</h2>
                            <p style="margin: 0 0 24px; font-size: 16px; line-height: 1.6; color: #a1a1aa;">
                                Мы получили запрос на сброс пароля для вашего аккаунта. Нажмите на кнопку ниже, чтобы создать новый пароль.
                            </p>

                            <!-- Button -->
                            <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0">
                                <tr>
                                    <td align="center" style="padding: 8px 0 24px;">
                                        <a href="%s" style="display: inline-block; padding: 14px 32px; background: linear-gradient(135deg, #8b5cf6 0%%, #a78bfa 100%%); color: #ffffff; text-decoration: none; font-size: 16px; font-weight: 600; border-radius: 12px; box-shadow: 0 4px 14px rgba(139, 92, 246, 0.4);">
                                            Сбросить пароль
                                        </a>
                                    </td>
                                </tr>
                            </table>

                            <p style="margin: 0 0 16px; font-size: 14px; line-height: 1.6; color: #71717a;">
                                Если кнопка не работает, скопируйте и вставьте эту ссылку в браузер:
                            </p>
                            <p style="margin: 0 0 24px; font-size: 14px; word-break: break-all; color: #8b5cf6;">
                                %s
                            </p>

                            <div style="padding: 16px; background-color: rgba(251, 191, 36, 0.1); border-radius: 8px; border: 1px solid rgba(251, 191, 36, 0.2);">
                                <p style="margin: 0; font-size: 14px; color: #fbbf24;">
                                    <strong>Важно:</strong> Ссылка действительна в течение 1 часа. Если вы не запрашивали сброс пароля, проигнорируйте это письмо.
                                </p>
                            </div>
                        </td>
                    </tr>

                    <!-- Footer -->
                    <tr>
                        <td style="padding: 30px 40px 40px;">
                            <hr style="border: none; border-top: 1px solid rgba(255, 255, 255, 0.1); margin: 0 0 20px;">
                            <p style="margin: 0; font-size: 13px; color: #52525b; text-align: center;">
                                Это автоматическое сообщение. Пожалуйста, не отвечайте на него.
                            </p>
                            <p style="margin: 8px 0 0; font-size: 13px; color: #52525b; text-align: center;">
                                &copy; 2025 Teamly. Все права защищены.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>`, resetURL, resetURL)
}

// GetPasswordResetEmailText возвращает текстовую версию письма для сброса пароля
func GetPasswordResetEmailText(resetURL string) string {
	return fmt.Sprintf(`Сброс пароля - Teamly

Мы получили запрос на сброс пароля для вашего аккаунта.

Чтобы создать новый пароль, перейдите по ссылке:
%s

Важно: Ссылка действительна в течение 1 часа. Если вы не запрашивали сброс пароля, проигнорируйте это письмо.

---
Это автоматическое сообщение. Пожалуйста, не отвечайте на него.
© 2025 Teamly. Все права защищены.`, resetURL)
}
