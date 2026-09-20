using System.Net;
using Microsoft.Extensions.Logging.Abstractions;
using Microsoft.Extensions.Options;
using TenantlyNotificationService.Configuration;
using TenantlyNotificationService.Services;

namespace TenantlyNotificationService.Tests;

public class SslWirelessSmsServiceTests
{
    private static NotificationSettings SettingsWithApiKey(string apiKey = "test-key") => new()
    {
        Sms = new SmsSettings
        {
            Provider = "ssl_wireless",
            ApiKey = apiKey,
            ApiUrl = "https://smsplus.sslwireless.com/api/v3/send-sms",
            SenderId = "TENANTLY",
        },
    };

    private static SslWirelessSmsService CreateService(NotificationSettings settings, HttpResponseMessage response)
    {
        var handler = new StubHttpMessageHandler(response);
        var httpClient = new HttpClient(handler);
        return new SslWirelessSmsService(Options.Create(settings), httpClient, NullLogger<SslWirelessSmsService>.Instance);
    }

    [Fact]
    public async Task SendSmsAsync_ReturnsTrue_WhenProviderRespondsSuccess()
    {
        var response = new HttpResponseMessage(HttpStatusCode.OK)
        {
            Content = new StringContent("""{"status":"SUCCESS","status_code":200}"""),
        };
        var service = CreateService(SettingsWithApiKey(), response);

        var result = await service.SendSmsAsync("+8801700000000", "Test message");

        Assert.True(result);
    }

    [Fact]
    public async Task SendSmsAsync_ReturnsFalse_WhenProviderRespondsFailureStatus()
    {
        var response = new HttpResponseMessage(HttpStatusCode.OK)
        {
            Content = new StringContent("""{"status":"FAILED","status_code":400,"error_message":"Invalid msisdn"}"""),
        };
        var service = CreateService(SettingsWithApiKey(), response);

        var result = await service.SendSmsAsync("+8801700000000", "Test message");

        Assert.False(result);
    }

    [Fact]
    public async Task SendSmsAsync_ReturnsFalse_WhenHttpStatusIsNotSuccess()
    {
        var response = new HttpResponseMessage(HttpStatusCode.InternalServerError)
        {
            Content = new StringContent("upstream error"),
        };
        var service = CreateService(SettingsWithApiKey(), response);

        var result = await service.SendSmsAsync("+8801700000000", "Test message");

        Assert.False(result);
    }

    [Fact]
    public async Task SendSmsAsync_ReturnsFalse_WhenResponseBodyIsMalformed()
    {
        var response = new HttpResponseMessage(HttpStatusCode.OK)
        {
            Content = new StringContent("not json"),
        };
        var service = CreateService(SettingsWithApiKey(), response);

        var result = await service.SendSmsAsync("+8801700000000", "Test message");

        Assert.False(result);
    }

    [Fact]
    public async Task SendSmsAsync_ReturnsFalse_WhenApiKeyIsNotConfigured()
    {
        var response = new HttpResponseMessage(HttpStatusCode.OK)
        {
            Content = new StringContent("""{"status":"SUCCESS"}"""),
        };
        var service = CreateService(SettingsWithApiKey(apiKey: string.Empty), response);

        var result = await service.SendSmsAsync("+8801700000000", "Test message");

        Assert.False(result);
    }

    private class StubHttpMessageHandler(HttpResponseMessage response) : HttpMessageHandler
    {
        protected override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken)
            => Task.FromResult(response);
    }
}
