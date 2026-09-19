using Microsoft.Extensions.DependencyInjection;

namespace TenantlyNotificationService.Services;

/// <summary>
/// Resolves the configured SMS provider name (Notifications:Sms:Provider) to
/// a concrete ISmsService implementation. ISmsService is the only contract
/// NotificationWorker depends on, so switching providers to compare delivery
/// cost/reliability is a config change, not a code change — as long as the
/// target provider is already registered here. Adding a brand new provider
/// means: implement ISmsService, register the class in DI (Program.cs), add
/// one case below.
/// </summary>
public static class SmsProviderFactory
{
    public const string SslWireless = "ssl_wireless";

    public static ISmsService Resolve(string? providerName, IServiceProvider services)
    {
        var normalized = string.IsNullOrWhiteSpace(providerName)
            ? SslWireless
            : providerName.Trim().ToLowerInvariant();

        return normalized switch
        {
            SslWireless => services.GetRequiredService<SslWirelessSmsService>(),
            _ => throw new InvalidOperationException(
                $"Unknown SMS provider '{providerName}' configured at Notifications:Sms:Provider"),
        };
    }
}
