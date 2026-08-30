using Microsoft.EntityFrameworkCore;
using TenantlyNotificationService.Models;
using TenantlyNotificationService.Models.Entities;
using TenantlyNotificationService.Models.Enums;

namespace TenantlyNotificationService.Data;

public class TenantlyDbContext : DbContext
{
    public TenantlyDbContext(DbContextOptions<TenantlyDbContext> options) : base(options)
    {
    }

    public DbSet<User> Users { get; set; }
    public DbSet<Property> Properties { get; set; }
    public DbSet<Unit> Units { get; set; }
    public DbSet<Tenant> Tenants { get; set; }
    public DbSet<Lease> Leases { get; set; }
    public DbSet<Payment> Payments { get; set; }
    public DbSet<NotificationQueue> NotificationQueue { get; set; }

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        base.OnModelCreating(modelBuilder);

        // Configure User entity
        modelBuilder.Entity<User>(entity =>
        {
            entity.HasIndex(e => e.Username).IsUnique();
            entity.HasIndex(e => e.Email).IsUnique();
        });

        // Configure Property entity
        modelBuilder.Entity<Property>(entity =>
        {
            entity.HasIndex(e => e.Active);
        });

        // Configure Unit entity
        modelBuilder.Entity<Unit>(entity =>
        {
            entity.HasIndex(e => e.Active);
            entity.HasIndex(e => new { e.PropertyId, e.Active });

            entity.HasOne(d => d.Property)
                .WithMany(p => p.Units)
                .HasForeignKey(d => d.PropertyId)
                .OnDelete(DeleteBehavior.Restrict);
        });

        // Configure Tenant entity
        modelBuilder.Entity<Tenant>(entity =>
        {
            entity.HasIndex(e => e.Active);
            entity.HasIndex(e => e.PhoneNumber).HasFilter("phone_number IS NOT NULL");
            entity.HasIndex(e => e.Email).HasFilter("email IS NOT NULL");
        });

        // Configure Lease entity
        modelBuilder.Entity<Lease>(entity =>
        {
            entity.HasIndex(e => e.Active);
            entity.HasIndex(e => new { e.UnitId, e.TenantId });
            entity.HasIndex(e => e.StartDate);
            entity.HasIndex(e => new { e.TenantId, e.Active });
            entity.HasIndex(e => new { e.UnitId, e.Active });

            entity.HasOne(d => d.Unit)
                .WithMany(p => p.Leases)
                .HasForeignKey(d => d.UnitId)
                .OnDelete(DeleteBehavior.Restrict);

            entity.HasOne(d => d.Tenant)
                .WithMany(p => p.Leases)
                .HasForeignKey(d => d.TenantId)
                .OnDelete(DeleteBehavior.Restrict);
        });

        // Configure Payment entity
        modelBuilder.Entity<Payment>(entity =>
        {
            entity.HasIndex(e => new { e.UnitId, e.Month, e.Year }).IsUnique();
            entity.HasIndex(e => e.Status);
            entity.HasIndex(e => e.DueDate);
            entity.HasIndex(e => new { e.TenantId, e.Status });
            entity.HasIndex(e => new { e.Year, e.Month });
            entity.HasIndex(e => e.ReceiptNumber).HasFilter("receipt_number IS NOT NULL");

            // Configure enum to be stored as string in database
            entity.Property(e => e.Status)
                .HasConversion<string>()
                .HasMaxLength(20);

            entity.HasOne(d => d.Unit)
                .WithMany(p => p.Payments)
                .HasForeignKey(d => d.UnitId)
                .OnDelete(DeleteBehavior.Restrict);

            entity.HasOne(d => d.Tenant)
                .WithMany(p => p.Payments)
                .HasForeignKey(d => d.TenantId)
                .OnDelete(DeleteBehavior.Restrict);
        });

        // Configure NotificationQueue entity
        modelBuilder.Entity<NotificationQueue>(entity =>
        {
            entity.HasIndex(e => e.Status);
            entity.HasIndex(e => e.CreatedAt);
            entity.HasIndex(e => new { e.TenantId, e.NotificationType });
            entity.HasIndex(e => new { e.Status, e.CreatedAt });

            entity.HasOne(d => d.Unit)
                .WithMany(p => p.Notifications)
                .HasForeignKey(d => d.UnitId)
                .OnDelete(DeleteBehavior.Restrict);

            entity.HasOne(d => d.Tenant)
                .WithMany(p => p.Notifications)
                .HasForeignKey(d => d.TenantId)
                .OnDelete(DeleteBehavior.Restrict);
        });
    }
}