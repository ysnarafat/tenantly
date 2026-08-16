import 'package:drift/drift.dart';

/// Mirrors the property types used on the backend/web app so a future sync
/// layer maps 1:1 onto the existing schema without a data-model rework.
enum PropertyType { residential, commercial, mixed }

enum BuildingType { residential, commercial, mixed }

enum UnitType { shop, apartment, office, parking, storage, other }

enum PaymentStatus { due, partial, paid, overdue }

@DataClassName('PropertyRow')
class Properties extends Table {
  IntColumn get id => integer().autoIncrement()();
  TextColumn get name => text().withLength(min: 1, max: 200)();
  TextColumn get code => text().withLength(min: 1, max: 50)();
  TextColumn get propertyType => textEnum<PropertyType>()();
  TextColumn get address => text()();
  DateTimeColumn get createdAt => dateTime().withDefault(currentDateAndTime)();
  DateTimeColumn get updatedAt => dateTime().withDefault(currentDateAndTime)();
}

@DataClassName('BuildingRow')
class Buildings extends Table {
  IntColumn get id => integer().autoIncrement()();
  IntColumn get propertyId =>
      integer().references(Properties, #id, onDelete: KeyAction.cascade)();
  TextColumn get name => text().withLength(min: 1, max: 200)();
  TextColumn get code => text().withLength(min: 1, max: 50)();
  TextColumn get buildingType => textEnum<BuildingType>()();
  DateTimeColumn get createdAt => dateTime().withDefault(currentDateAndTime)();
  DateTimeColumn get updatedAt => dateTime().withDefault(currentDateAndTime)();
}

@DataClassName('UnitRow')
class Units extends Table {
  IntColumn get id => integer().autoIncrement()();
  IntColumn get buildingId =>
      integer().references(Buildings, #id, onDelete: KeyAction.cascade)();
  TextColumn get unitNumber => text().withLength(min: 1, max: 50)();
  TextColumn get unitType => textEnum<UnitType>()();
  IntColumn get floor => integer().nullable()();
  DateTimeColumn get createdAt => dateTime().withDefault(currentDateAndTime)();
  DateTimeColumn get updatedAt => dateTime().withDefault(currentDateAndTime)();
}

/// NID/phone are stored in plaintext for now — there is no on-device
/// encryption-at-rest infrastructure yet. Flagged explicitly (see
/// docs/AI_INTEGRATION_ROADMAP... — actually see the mobile setup plan) as a
/// follow-up once sync/security requirements for this app are decided, rather
/// than silently under-protecting it.
@DataClassName('TenantRow')
class Tenants extends Table {
  IntColumn get id => integer().autoIncrement()();
  TextColumn get name => text().withLength(min: 1, max: 200)();
  TextColumn get phone => text().nullable()();
  TextColumn get nidNumber => text().nullable()();
  DateTimeColumn get createdAt => dateTime().withDefault(currentDateAndTime)();
  DateTimeColumn get updatedAt => dateTime().withDefault(currentDateAndTime)();
}

@DataClassName('LeaseRow')
class Leases extends Table {
  IntColumn get id => integer().autoIncrement()();
  IntColumn get unitId =>
      integer().references(Units, #id, onDelete: KeyAction.cascade)();
  IntColumn get tenantId =>
      integer().references(Tenants, #id, onDelete: KeyAction.cascade)();
  RealColumn get monthlyRent => real()();
  DateTimeColumn get startDate => dateTime()();
  DateTimeColumn get endDate => dateTime().nullable()();
  BoolColumn get active => boolean().withDefault(const Constant(true))();
  DateTimeColumn get createdAt => dateTime().withDefault(currentDateAndTime)();
  DateTimeColumn get updatedAt => dateTime().withDefault(currentDateAndTime)();
}

/// One row per rent period for a lease — mirrors the web app's Payment
/// model, but references the lease directly (unit/tenant/building/property
/// are all reachable through it) rather than denormalizing those ids here.
@DataClassName('PaymentRow')
class Payments extends Table {
  IntColumn get id => integer().autoIncrement()();
  IntColumn get leaseId =>
      integer().references(Leases, #id, onDelete: KeyAction.cascade)();
  IntColumn get month => integer()(); // 1-12
  IntColumn get year => integer()();
  RealColumn get amountDue => real()();
  RealColumn get amountPaid => real().withDefault(const Constant(0))();
  TextColumn get status => textEnum<PaymentStatus>()();
  TextColumn get paymentMethod => text().nullable()();
  DateTimeColumn get paymentDate => dateTime().nullable()();
  TextColumn get notes => text().nullable()();
  DateTimeColumn get createdAt => dateTime().withDefault(currentDateAndTime)();
  DateTimeColumn get updatedAt => dateTime().withDefault(currentDateAndTime)();
}
