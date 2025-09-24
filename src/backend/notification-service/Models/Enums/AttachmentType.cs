namespace TenantlyNotificationService.Models.Enums;

public enum AttachmentType
{
    // Payment related attachments
    PaymentReceipt = 1,
    BankSlip = 2,
    PaymentVoucher = 3,
    CashMemo = 4,
    
    // Property related attachments
    PropertyDeed = 10,
    OwnershipCertificate = 11,
    TradeLicense = 12,
    BuildingPermit = 13,
    PropertyTaxReceipt = 14,
    UtilityBill = 15,
    PropertyPhoto = 16,
    FloorPlan = 17,
    
    // Lease related attachments
    LeaseAgreement = 20,
    RentAgreement = 21,
    SecurityDepositReceipt = 22,
    LeaseRenewal = 23,
    EvictionNotice = 24,
    
    // Shop related attachments
    ShopPhoto = 30,
    ShopLayout = 31,
    ShopInventory = 32,
    MaintenanceRecord = 33,
    
    // Tenant related attachments
    TenantNID = 40,
    TenantPhoto = 41,
    BusinessLicense = 42,
    TenantReference = 43,
    
    // General attachments
    Contract = 50,
    Invoice = 51,
    Image = 60,
    PDF = 61,
    Spreadsheet = 62,
    TextDocument = 63,
    Other = 99
}

public enum AttachmentStatus
{
    Active = 1,
    Archived = 2,
    Deleted = 3
}