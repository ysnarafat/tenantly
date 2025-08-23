namespace TenantlyNotificationService.Models;

public class OverduePayment
{
    public int Id { get; set; }
    public int ShopId { get; set; }
    public int TenantId { get; set; }
    public int Month { get; set; }
    public int Year { get; set; }
    public decimal AmountDue { get; set; }
    public string ShopName { get; set; } = string.Empty;
    public string TenantName { get; set; } = string.Empty;
    public string? PhoneNumber { get; set; }
}