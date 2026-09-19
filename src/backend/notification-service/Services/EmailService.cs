using MailKit.Net.Smtp;
using MailKit.Security;
using Microsoft.Extensions.Options;
using MimeKit;
using TenantlyNotificationService.Configuration;

namespace TenantlyNotificationService.Services;

public class EmailService : IEmailService
{
    private readonly NotificationSettings _settings;
    private readonly ILogger<EmailService> _logger;
    private bool _missingCredentialsWarningLogged;

    public EmailService(IOptions<NotificationSettings> settings, ILogger<EmailService> logger)
    {
        _settings = settings.Value;
        _logger = logger;
    }

    public async Task<bool> SendEmailAsync(string toEmail, string subject, string body)
    {
        var email = _settings.Email;

        if (string.IsNullOrEmpty(email.Username))
        {
            if (!_missingCredentialsWarningLogged)
            {
                _logger.LogWarning("SMTP Username is not configured; email sending is disabled until Notifications:Email:Username/Password are set");
                _missingCredentialsWarningLogged = true;
            }
            return false;
        }

        try
        {
            var message = new MimeMessage();
            message.From.Add(new MailboxAddress(email.FromName, email.FromEmail));
            message.To.Add(MailboxAddress.Parse(toEmail));
            message.Subject = subject;
            message.Body = new TextPart("plain") { Text = body };

            using var client = new SmtpClient();
            var secureOption = email.EnableSsl ? SecureSocketOptions.StartTls : SecureSocketOptions.None;
            await client.ConnectAsync(email.SmtpHost, email.SmtpPort, secureOption);
            await client.AuthenticateAsync(email.Username, email.Password);
            await client.SendAsync(message);
            await client.DisconnectAsync(true);

            _logger.LogInformation("Email sent to {Email}: {Subject}", toEmail, subject);
            return true;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error sending email to {Email}", toEmail);
            return false;
        }
    }
}
