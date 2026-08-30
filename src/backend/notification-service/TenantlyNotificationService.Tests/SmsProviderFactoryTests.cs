using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Logging.Abstractions;
using Microsoft.Extensions.Options;
using TenantlyNotificationService.Configuration;
using TenantlyNotificationService.Services;

namespace TenantlyNotificationService.Tests;

public class SmsProviderFactoryTests
{
    private static IServiceProvider BuildServiceProvider()
    {
        var services = new ServiceCollection();
        services.AddSingleton(Options.Create(new NotificationSettings()));
        services.AddSingleton(new HttpClient());
        services.AddSingleton<ILogger<SslWirelessSmsService>>(NullLogger<SslWirelessSmsService>.Instance);
        services.AddSingleton<SslWirelessSmsService>();
        return services.BuildServiceProvider();
    }

    [Fact]
    public void Resolve_ReturnsSslWirelessProvider_WhenExplicitlyConfigured()
    {
        var sp = BuildServiceProvider();

        var result = SmsProviderFactory.Resolve("ssl_wireless", sp);

        Assert.IsType<SslWirelessSmsService>(result);
    }

    [Theory]
    [InlineData(null)]
    [InlineData("")]
    [InlineData("   ")]
    public void Resolve_DefaultsToSslWireless_WhenProviderNotConfigured(string? provider)
    {
        var sp = BuildServiceProvider();

        var result = SmsProviderFactory.Resolve(provider, sp);

        Assert.IsType<SslWirelessSmsService>(result);
    }

    [Fact]
    public void Resolve_IsCaseInsensitive()
    {
        var sp = BuildServiceProvider();

        var result = SmsProviderFactory.Resolve("SSL_Wireless", sp);

        Assert.IsType<SslWirelessSmsService>(result);
    }

    [Fact]
    public void Resolve_ThrowsForUnknownProvider()
    {
        var sp = BuildServiceProvider();

        var ex = Assert.Throws<InvalidOperationException>(() => SmsProviderFactory.Resolve("twilio", sp));

        Assert.Contains("twilio", ex.Message);
    }
}
