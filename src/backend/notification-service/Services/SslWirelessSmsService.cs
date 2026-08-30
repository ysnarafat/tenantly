using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;
using Microsoft.Extensions.Options;
using TenantlyNotificationService.Configuration;

namespace TenantlyNotificationService.Services;

// SSL Wireless is the current SMS gateway for Bangladesh delivery. This class
// holds all of its provider-specific request/response shape; ISmsService is
// the only contract callers (NotificationWorker) depend on, so a future
// provider swap means adding a new ISmsService implementation and a case in
// SmsProviderFactory — not touching this class or its callers.
public class SslWirelessSmsService : ISmsService
{
    private readonly NotificationSettings _settings;
    private readonly HttpClient _httpClient;
    private readonly ILogger<SslWirelessSmsService> _logger;
    private bool _missingCredentialsWarningLogged;

    public SslWirelessSmsService(IOptions<NotificationSettings> settings, HttpClient httpClient, ILogger<SslWirelessSmsService> logger)
    {
        _settings = settings.Value;
        _httpClient = httpClient;
        _logger = logger;
    }

    public async Task<bool> SendSmsAsync(string phoneNumber, string message)
    {
        var sms = _settings.Sms;

        if (string.IsNullOrEmpty(sms.ApiKey))
        {
            if (!_missingCredentialsWarningLogged)
            {
                _logger.LogWarning("SMS ApiKey is not configured; SMS sending is disabled until Notifications:Sms:ApiKey is set");
                _missingCredentialsWarningLogged = true;
            }
            return false;
        }

        try
        {
            var payload = new SslWirelessSendSmsRequest
            {
                ApiToken = sms.ApiKey,
                Sid = sms.SenderId,
                Msisdn = phoneNumber,
                Sms = message,
                CsmsId = Guid.NewGuid().ToString("N"),
            };

            using var content = new StringContent(JsonSerializer.Serialize(payload), Encoding.UTF8, "application/json");
            using var response = await _httpClient.PostAsync(sms.ApiUrl, content);
            var responseBody = await response.Content.ReadAsStringAsync();

            if (!response.IsSuccessStatusCode)
            {
                _logger.LogError("SMS provider returned HTTP {StatusCode} for {PhoneNumber}: {Body}", (int)response.StatusCode, phoneNumber, responseBody);
                return false;
            }

            var result = JsonSerializer.Deserialize<SslWirelessSendSmsResponse>(responseBody, JsonOptions);
            if (result == null || !string.Equals(result.Status, "SUCCESS", StringComparison.OrdinalIgnoreCase))
            {
                _logger.LogError("SMS delivery failed for {PhoneNumber}: {Body}", phoneNumber, responseBody);
                return false;
            }

            _logger.LogInformation("SMS sent to {PhoneNumber}", phoneNumber);
            return true;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error sending SMS to {PhoneNumber}", phoneNumber);
            return false;
        }
    }

    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNameCaseInsensitive = true,
    };

    private class SslWirelessSendSmsRequest
    {
        [JsonPropertyName("api_token")]
        public string ApiToken { get; set; } = string.Empty;

        [JsonPropertyName("sid")]
        public string Sid { get; set; } = string.Empty;

        [JsonPropertyName("msisdn")]
        public string Msisdn { get; set; } = string.Empty;

        [JsonPropertyName("sms")]
        public string Sms { get; set; } = string.Empty;

        [JsonPropertyName("csms_id")]
        public string CsmsId { get; set; } = string.Empty;
    }

    private class SslWirelessSendSmsResponse
    {
        [JsonPropertyName("status")]
        public string? Status { get; set; }

        [JsonPropertyName("status_code")]
        public int? StatusCode { get; set; }

        [JsonPropertyName("error_message")]
        public string? ErrorMessage { get; set; }
    }
}
