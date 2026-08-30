using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Options;
using Serilog;
using TenantlyNotificationService.Configuration;
using TenantlyNotificationService.Data;
using TenantlyNotificationService.Data.Interceptors;
using TenantlyNotificationService.Services;

var builder = Host.CreateApplicationBuilder(args);

// Configure Serilog
Log.Logger = new LoggerConfiguration()
    .ReadFrom.Configuration(builder.Configuration)
    .CreateLogger();

builder.Services.AddSerilog();

// Add EF Core
var connectionString = ConnectionStringResolver.Resolve(builder.Configuration);

// Register the interceptor as a service
builder.Services.AddScoped<AuditInterceptor>();

builder.Services.AddDbContext<TenantlyDbContext>((serviceProvider, options) =>
    options.UseNpgsql(connectionString)
           .AddInterceptors(serviceProvider.GetRequiredService<AuditInterceptor>()));

// Add configuration
builder.Services.Configure<NotificationSettings>(
    builder.Configuration.GetSection("Notifications"));

// Add HTTP client
builder.Services.AddHttpClient();

// Add services
builder.Services.AddScoped<IDatabaseService, DatabaseService>();

// SMS provider is selected by Notifications:Sms:Provider — see
// SmsProviderFactory for how to add a new provider.
builder.Services.AddSingleton<SslWirelessSmsService>();
builder.Services.AddSingleton<ISmsService>(sp => SmsProviderFactory.Resolve(
    sp.GetRequiredService<IOptions<NotificationSettings>>().Value.Sms.Provider, sp));
builder.Services.AddSingleton<IEmailService, EmailService>();
builder.Services.AddSingleton<INotificationChannelListener, PostgresNotificationChannelListener>();

// Add hosted services
builder.Services.AddHostedService<NotificationWorker>();
builder.Services.AddHostedService<ReminderWorker>();

var host = builder.Build();

try
{
    Log.Information("Starting Tenantly Notification Service");

    await host.RunAsync();
}
catch (Exception ex)
{
    Log.Fatal(ex, "Notification service terminated unexpectedly");
}
finally
{
    Log.CloseAndFlush();
}