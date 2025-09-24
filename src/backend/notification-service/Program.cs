using TenantlyNotificationService.Services;
using TenantlyNotificationService.Configuration;
using TenantlyNotificationService.Data;
using TenantlyNotificationService.Data.Interceptors;
using Microsoft.EntityFrameworkCore;
using Serilog;

var builder = Host.CreateApplicationBuilder(args);

// Configure Serilog
Log.Logger = new LoggerConfiguration()
    .ReadFrom.Configuration(builder.Configuration)
    .CreateLogger();

builder.Services.AddSerilog();

// Add EF Core
var connectionString = builder.Configuration.GetConnectionString("DefaultConnection") 
    ?? builder.Configuration.GetSection("Database:ConnectionString").Value
    ?? throw new InvalidOperationException("Connection string not found.");

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
builder.Services.AddSingleton<ISmsService, SmsService>();
builder.Services.AddSingleton<IEmailService, EmailService>();

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