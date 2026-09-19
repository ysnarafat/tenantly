namespace TenantlyNotificationService.Configuration;

public class NotificationSettings
{
    public SmsSettings Sms { get; set; } = new();
    public EmailSettings Email { get; set; } = new();
    public int ProcessingIntervalSeconds { get; set; } = 30;
    // A notification is retried on subsequent poll cycles until it either
    // succeeds or exhausts this many attempts, after which it is left
    // permanently Failed rather than retried indefinitely.
    public int MaxRetryAttempts { get; set; } = 5;
}

public class SmsSettings
{
    public string Provider { get; set; } = string.Empty;
    public string ApiKey { get; set; } = string.Empty;
    public string ApiUrl { get; set; } = string.Empty;
    public string SenderId { get; set; } = string.Empty;
}

public class EmailSettings
{
    public string SmtpHost { get; set; } = string.Empty;
    public int SmtpPort { get; set; } = 587;
    public string Username { get; set; } = string.Empty;
    public string Password { get; set; } = string.Empty;
    public string FromEmail { get; set; } = string.Empty;
    public string FromName { get; set; } = "Tenantly";
    public bool EnableSsl { get; set; } = true;
}