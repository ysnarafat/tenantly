using TenantlyNotificationService.Models.Enums;

namespace TenantlyNotificationService.Extensions;

public static class PaymentStatusExtensions
{
    public static string ToDisplayString(this PaymentStatus status)
    {
        return status switch
        {
            PaymentStatus.Due => "Due",
            PaymentStatus.Paid => "Paid",
            PaymentStatus.Partial => "Partial",
            _ => status.ToString()
        };
    }

    public static PaymentStatus FromString(string status)
    {
        return status?.ToLowerInvariant() switch
        {
            "due" => PaymentStatus.Due,
            "paid" => PaymentStatus.Paid,
            "partial" => PaymentStatus.Partial,
            _ => PaymentStatus.Due
        };
    }

    public static bool IsOverdue(this PaymentStatus status)
    {
        return status == PaymentStatus.Due || status == PaymentStatus.Partial;
    }

    public static bool IsFullyPaid(this PaymentStatus status)
    {
        return status == PaymentStatus.Paid;
    }
}