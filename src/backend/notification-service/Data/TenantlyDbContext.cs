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
    public DbSet<Shop> Shops { get; set; }
    public DbSet<Tenant> Tenants { get; set; }
    public DbSet<Lease> Leases { get; set; }
    public DbSet<Payment> Payments { get; set; }
    public DbSet<NotificationQueue> NotificationQueue { get; set; }
    public DbSet<Attachment> Attachments { get; set; }

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

        // Configure Shop entity
        modelBuilder.Entity<Shop>(entity =>
        {
            entity.HasIndex(e => e.Active);
            entity.HasIndex(e => new { e.PropertyId, e.Active });
            entity.Property(e => e.PropertyId).HasDefaultValue(1);

            entity.HasOne(d => d.Property)
                .WithMany(p => p.Shops)
                .HasForeignKey(d => d.PropertyId)
                .OnDelete(DeleteBehavior.Restrict);
        });

        // Configure Tenant entity
        modelBuilder.Entity<Tenant>(entity =>
        {
            entity.HasIndex(e => e.Active);
            entity.HasIndex(e => e.PhoneNumber).HasFilter("phone_number IS NOT NULL");
            entity.HasIndex(e => e.Email).HasFilter("email IS NOT NULL");
            entity.HasIndex(e => e.NidNumber).HasFilter("nid_number IS NOT NULL");
        });

        // Configure Lease entity
        modelBuilder.Entity<Lease>(entity =>
        {
            entity.HasIndex(e => e.Active);
            entity.HasIndex(e => new { e.ShopId, e.TenantId });
            entity.HasIndex(e => e.StartDate);
            entity.HasIndex(e => new { e.TenantId, e.Active });
            entity.HasIndex(e => new { e.ShopId, e.Active });

            entity.HasOne(d => d.Shop)
                .WithMany(p => p.Leases)
                .HasForeignKey(d => d.ShopId)
                .OnDelete(DeleteBehavior.Restrict);

            entity.HasOne(d => d.Tenant)
                .WithMany(p => p.Leases)
                .HasForeignKey(d => d.TenantId)
                .OnDelete(DeleteBehavior.Restrict);
        });

        // Configure Payment entity
        modelBuilder.Entity<Payment>(entity =>
        {
            entity.HasIndex(e => new { e.ShopId, e.Month, e.Year }).IsUnique();
            entity.HasIndex(e => e.Status);
            entity.HasIndex(e => e.DueDate);
            entity.HasIndex(e => new { e.TenantId, e.Status });
            entity.HasIndex(e => new { e.Year, e.Month });
            entity.HasIndex(e => e.ReceiptNumber).HasFilter("receipt_number IS NOT NULL");
            
            // Configure enum to be stored as string in database
            entity.Property(e => e.Status)
                .HasConversion<string>()
                .HasMaxLength(20);

            entity.HasOne(d => d.Shop)
                .WithMany(p => p.Payments)
                .HasForeignKey(d => d.ShopId)
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

            entity.HasOne(d => d.Shop)
                .WithMany(p => p.Notifications)
                .HasForeignKey(d => d.ShopId)
                .OnDelete(DeleteBehavior.Restrict);

            entity.HasOne(d => d.Tenant)
                .WithMany(p => p.Notifications)
                .HasForeignKey(d => d.TenantId)
                .OnDelete(DeleteBehavior.Restrict);
        });

        // Configure Attachment entity
        modelBuilder.Entity<Attachment>(entity =>
        {
            entity.HasIndex(e => e.AttachmentType);
            entity.HasIndex(e => e.Status);
            entity.HasIndex(e => e.CreatedAt);
            entity.HasIndex(e => new { e.PaymentId, e.AttachmentType }).HasFilter("payment_id IS NOT NULL");
            entity.HasIndex(e => new { e.PropertyId, e.AttachmentType }).HasFilter("property_id IS NOT NULL");
            entity.HasIndex(e => new { e.ShopId, e.AttachmentType }).HasFilter("shop_id IS NOT NULL");
            entity.HasIndex(e => new { e.LeaseId, e.AttachmentType }).HasFilter("lease_id IS NOT NULL");
            entity.HasIndex(e => new { e.TenantId, e.AttachmentType }).HasFilter("tenant_id IS NOT NULL");
            
            // Configure enums to be stored as strings
            entity.Property(e => e.AttachmentType)
                .HasConversion<string>()
                .HasMaxLength(50);
            
            entity.Property(e => e.Status)
                .HasConversion<string>()
                .HasMaxLength(20);

            // Configure optional foreign key relationships
            entity.HasOne(d => d.Payment)
                .WithMany(p => p.Attachments)
                .HasForeignKey(d => d.PaymentId)
                .OnDelete(DeleteBehavior.Cascade);

            entity.HasOne(d => d.Property)
                .WithMany(p => p.Attachments)
                .HasForeignKey(d => d.PropertyId)
                .OnDelete(DeleteBehavior.Cascade);

            entity.HasOne(d => d.Shop)
                .WithMany(s => s.Attachments)
                .HasForeignKey(d => d.ShopId)
                .OnDelete(DeleteBehavior.Cascade);

            entity.HasOne(d => d.Lease)
                .WithMany(l => l.Attachments)
                .HasForeignKey(d => d.LeaseId)
                .OnDelete(DeleteBehavior.Cascade);

            entity.HasOne(d => d.Tenant)
                .WithMany(t => t.Attachments)
                .HasForeignKey(d => d.TenantId)
                .OnDelete(DeleteBehavior.Cascade);

            entity.HasOne(d => d.UploadedByUser)
                .WithMany()
                .HasForeignKey(d => d.UploadedBy)
                .OnDelete(DeleteBehavior.SetNull);

            // Add constraint to ensure only one entity relationship is set
            entity.HasCheckConstraint("CK_Attachment_SingleEntity", 
                "(CASE WHEN payment_id IS NOT NULL THEN 1 ELSE 0 END + " +
                "CASE WHEN property_id IS NOT NULL THEN 1 ELSE 0 END + " +
                "CASE WHEN shop_id IS NOT NULL THEN 1 ELSE 0 END + " +
                "CASE WHEN lease_id IS NOT NULL THEN 1 ELSE 0 END + " +
                "CASE WHEN tenant_id IS NOT NULL THEN 1 ELSE 0 END) = 1");
        });
    }
}