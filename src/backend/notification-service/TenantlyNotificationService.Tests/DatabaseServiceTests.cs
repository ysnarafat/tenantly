using Microsoft.Data.Sqlite;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Logging.Abstractions;
using Microsoft.Extensions.Options;
using TenantlyNotificationService.Configuration;
using TenantlyNotificationService.Data;
using TenantlyNotificationService.Models;
using TenantlyNotificationService.Models.Entities;
using TenantlyNotificationService.Models.Enums;
using TenantlyNotificationService.Services;

namespace TenantlyNotificationService.Tests;

public class DatabaseServiceTests : IDisposable
{
    private readonly SqliteConnection _connection;
    private readonly TenantlyDbContext _context;
    private readonly DatabaseService _service;

    public DatabaseServiceTests()
    {
        _connection = new SqliteConnection("DataSource=:memory:");
        _connection.Open();

        var options = new DbContextOptionsBuilder<TenantlyDbContext>()
            .UseSqlite(_connection)
            .Options;

        _context = new TenantlyDbContext(options);
        _context.Database.EnsureCreated();

        _service = new DatabaseService(_context, Options.Create(new NotificationSettings()), NullLogger<DatabaseService>.Instance);
    }

    public void Dispose()
    {
        _context.Dispose();
        _connection.Dispose();
    }

    private async Task<(Tenant tenant, Unit unit)> SeedTenantAndUnitAsync()
    {
        var property = new Property { PropertyName = "Test Property", PropertyCode = "TP1", Address = "123 Road", CreatedAt = DateTime.UtcNow, UpdatedAt = DateTime.UtcNow };
        var unit = new Unit { UnitNumber = "U-1", UnitName = "Shop 1", Property = property, CreatedAt = DateTime.UtcNow, UpdatedAt = DateTime.UtcNow };
        var tenant = new Tenant { Name = "Jane Doe", PhoneNumber = "+8801700000000", Email = "jane@example.com", CreatedAt = DateTime.UtcNow, UpdatedAt = DateTime.UtcNow };

        _context.Properties.Add(property);
        _context.Units.Add(unit);
        _context.Tenants.Add(tenant);
        await _context.SaveChangesAsync();

        return (tenant, unit);
    }

    [Fact]
    public async Task GetPendingNotificationsAsync_ReturnsOnlyPendingBelowMaxRetries()
    {
        var (tenant, unit) = await SeedTenantAndUnitAsync();

        _context.NotificationQueue.AddRange(
            new NotificationQueue { TenantId = tenant.Id, UnitId = unit.Id, Message = "pending", NotificationType = "SMS", Recipient = "x", Status = "Pending", RetryCount = 0, CreatedAt = DateTime.UtcNow, UpdatedAt = DateTime.UtcNow },
            new NotificationQueue { TenantId = tenant.Id, UnitId = unit.Id, Message = "exhausted", NotificationType = "SMS", Recipient = "x", Status = "Pending", RetryCount = 3, CreatedAt = DateTime.UtcNow, UpdatedAt = DateTime.UtcNow },
            new NotificationQueue { TenantId = tenant.Id, UnitId = unit.Id, Message = "sent", NotificationType = "SMS", Recipient = "x", Status = "Sent", RetryCount = 0, CreatedAt = DateTime.UtcNow, UpdatedAt = DateTime.UtcNow }
        );
        await _context.SaveChangesAsync();

        var pending = await _service.GetPendingNotificationsAsync();

        var pendingList = pending.ToList();
        Assert.Single(pendingList);
        Assert.Equal("pending", pendingList[0].Message);
    }

    [Fact]
    public async Task UpdateNotificationStatusAsync_SetsStatusAndIncrementsRetryCount()
    {
        var (tenant, unit) = await SeedTenantAndUnitAsync();
        var notification = new NotificationQueue { TenantId = tenant.Id, UnitId = unit.Id, Message = "m", NotificationType = "SMS", Recipient = "x", Status = "Pending", RetryCount = 0, CreatedAt = DateTime.UtcNow, UpdatedAt = DateTime.UtcNow };
        _context.NotificationQueue.Add(notification);
        await _context.SaveChangesAsync();

        await _service.UpdateNotificationStatusAsync(notification.Id, NotificationStatus.Failed, "delivery failed");

        var updated = await _context.NotificationQueue.FindAsync(notification.Id);
        Assert.NotNull(updated);
        Assert.Equal("Failed", updated!.Status);
        Assert.Equal(1, updated.RetryCount);
        Assert.Equal("delivery failed", updated.ErrorMessage);
    }

    [Fact]
    public async Task GetOverduePaymentsAsync_ExcludesTenantsRemindedInLastDay()
    {
        var (tenant, unit) = await SeedTenantAndUnitAsync();
        var oldDueDate = DateOnly.FromDateTime(DateTime.UtcNow.AddDays(-10));

        var overduePayment = new Payment
        {
            UnitId = unit.Id,
            TenantId = tenant.Id,
            Month = 1,
            Year = 2026,
            AmountDue = 5000,
            Status = PaymentStatus.Due,
            DueDate = oldDueDate,
            CreatedAt = DateTime.UtcNow,
            UpdatedAt = DateTime.UtcNow,
        };
        _context.Payments.Add(overduePayment);
        await _context.SaveChangesAsync();

        _context.NotificationQueue.Add(new NotificationQueue
        {
            TenantId = tenant.Id,
            UnitId = unit.Id,
            Message = "already reminded",
            NotificationType = "Reminder",
            Recipient = tenant.PhoneNumber!,
            Status = "Sent",
            CreatedAt = DateTime.UtcNow.AddHours(-1),
            UpdatedAt = DateTime.UtcNow,
        });
        await _context.SaveChangesAsync();

        var overdue = await _service.GetOverduePaymentsAsync();

        Assert.Empty(overdue);
    }

    [Fact]
    public async Task GetOverduePaymentsAsync_ReturnsPaymentsNotYetReminded()
    {
        var (tenant, unit) = await SeedTenantAndUnitAsync();
        var oldDueDate = DateOnly.FromDateTime(DateTime.UtcNow.AddDays(-10));

        _context.Payments.Add(new Payment
        {
            UnitId = unit.Id,
            TenantId = tenant.Id,
            Month = 1,
            Year = 2026,
            AmountDue = 5000,
            Status = PaymentStatus.Due,
            DueDate = oldDueDate,
            CreatedAt = DateTime.UtcNow,
            UpdatedAt = DateTime.UtcNow,
        });
        await _context.SaveChangesAsync();

        var overdue = (await _service.GetOverduePaymentsAsync()).ToList();

        Assert.Single(overdue);
        Assert.Equal(tenant.Name, overdue[0].TenantName);
        Assert.Equal("Shop 1", overdue[0].UnitName);
    }
}
