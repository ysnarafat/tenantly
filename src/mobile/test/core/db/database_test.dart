import 'package:drift/drift.dart' show Value;
import 'package:drift/native.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tenantly_mobile/core/db/database.dart';
import 'package:tenantly_mobile/core/db/tables.dart';

/// Seeds a full property -> building -> unit -> tenant -> lease -> payment
/// chain and returns every id, so cascade/relationship tests don't each have
/// to re-derive this boilerplate.
class _SeededChain {
  _SeededChain({
    required this.propertyId,
    required this.buildingId,
    required this.unitId,
    required this.tenantId,
    required this.leaseId,
    required this.paymentId,
  });

  final int propertyId;
  final int buildingId;
  final int unitId;
  final int tenantId;
  final int leaseId;
  final int paymentId;
}

Future<_SeededChain> _seedFullChain(AppDatabase db) async {
  final propertyId = await db
      .into(db.properties)
      .insert(
        PropertiesCompanion.insert(
          name: 'Sunset Apartments',
          code: 'PROP-001',
          propertyType: PropertyType.residential,
          address: '123 Main St',
        ),
      );
  final buildingId = await db
      .into(db.buildings)
      .insert(
        BuildingsCompanion.insert(
          propertyId: propertyId,
          name: 'Tower A',
          code: 'BLD-001',
          buildingType: BuildingType.residential,
        ),
      );
  final unitId = await db
      .into(db.units)
      .insert(
        UnitsCompanion.insert(
          buildingId: buildingId,
          unitNumber: '101',
          unitType: UnitType.apartment,
        ),
      );
  final tenantId = await db
      .into(db.tenants)
      .insert(TenantsCompanion.insert(name: 'John Doe'));
  final leaseId = await db
      .into(db.leases)
      .insert(
        LeasesCompanion.insert(
          unitId: unitId,
          tenantId: tenantId,
          monthlyRent: 15000,
          startDate: DateTime(2026, 1, 1),
        ),
      );
  final paymentId = await db
      .into(db.payments)
      .insert(
        PaymentsCompanion.insert(
          leaseId: leaseId,
          month: 1,
          year: 2026,
          amountDue: 15000,
          status: PaymentStatus.due,
        ),
      );

  return _SeededChain(
    propertyId: propertyId,
    buildingId: buildingId,
    unitId: unitId,
    tenantId: tenantId,
    leaseId: leaseId,
    paymentId: paymentId,
  );
}

void main() {
  late AppDatabase db;

  setUp(() {
    db = AppDatabase.forTesting(NativeDatabase.memory());
  });

  tearDown(() async {
    await db.close();
  });

  test('inserts and queries a property', () async {
    final id = await db
        .into(db.properties)
        .insert(
          PropertiesCompanion.insert(
            name: 'Sunset Apartments',
            code: 'PROP-001',
            propertyType: PropertyType.residential,
            address: '123 Main St',
          ),
        );

    final rows = await db.select(db.properties).get();
    expect(rows, hasLength(1));
    expect(rows.single.id, id);
    expect(rows.single.name, 'Sunset Apartments');
    expect(rows.single.propertyType, PropertyType.residential);
  });

  test('property -> building -> unit foreign-key round-trip', () async {
    final propertyId = await db
        .into(db.properties)
        .insert(
          PropertiesCompanion.insert(
            name: 'Sunset Apartments',
            code: 'PROP-001',
            propertyType: PropertyType.residential,
            address: '123 Main St',
          ),
        );
    final buildingId = await db
        .into(db.buildings)
        .insert(
          BuildingsCompanion.insert(
            propertyId: propertyId,
            name: 'Tower A',
            code: 'BLD-001',
            buildingType: BuildingType.residential,
          ),
        );
    final unitId = await db
        .into(db.units)
        .insert(
          UnitsCompanion.insert(
            buildingId: buildingId,
            unitNumber: '101',
            unitType: UnitType.apartment,
          ),
        );

    final unit = await (db.select(
      db.units,
    )..where((t) => t.id.equals(unitId))).getSingle();
    expect(unit.buildingId, buildingId);

    final building = await (db.select(
      db.buildings,
    )..where((t) => t.id.equals(buildingId))).getSingle();
    expect(building.propertyId, propertyId);
  });

  test('deleting a property cascades to its buildings', () async {
    final propertyId = await db
        .into(db.properties)
        .insert(
          PropertiesCompanion.insert(
            name: 'Sunset Apartments',
            code: 'PROP-001',
            propertyType: PropertyType.residential,
            address: '123 Main St',
          ),
        );
    await db
        .into(db.buildings)
        .insert(
          BuildingsCompanion.insert(
            propertyId: propertyId,
            name: 'Tower A',
            code: 'BLD-001',
            buildingType: BuildingType.residential,
          ),
        );

    await (db.delete(
      db.properties,
    )..where((t) => t.id.equals(propertyId))).go();

    final remainingBuildings = await db.select(db.buildings).get();
    expect(remainingBuildings, isEmpty);
  });

  group('Tenants', () {
    test('inserts, queries, and updates a tenant', () async {
      final id = await db
          .into(db.tenants)
          .insert(
            TenantsCompanion.insert(
              name: 'Jane Doe',
              phone: const Value('01712345678'),
              nidNumber: const Value('1234567890'),
            ),
          );

      final tenant = await (db.select(
        db.tenants,
      )..where((t) => t.id.equals(id))).getSingle();
      expect(tenant.name, 'Jane Doe');
      expect(tenant.phone, '01712345678');
      expect(tenant.nidNumber, '1234567890');

      await db
          .update(db.tenants)
          .replace(tenant.copyWith(phone: const Value(null)));
      final updated = await (db.select(
        db.tenants,
      )..where((t) => t.id.equals(id))).getSingle();
      expect(updated.phone, isNull);
    });

    test('phone and NID are nullable', () async {
      final id = await db
          .into(db.tenants)
          .insert(TenantsCompanion.insert(name: 'No Contact Info'));
      final tenant = await (db.select(
        db.tenants,
      )..where((t) => t.id.equals(id))).getSingle();
      expect(tenant.phone, isNull);
      expect(tenant.nidNumber, isNull);
    });
  });

  group('Leases', () {
    test('lease -> unit and lease -> tenant foreign keys round-trip', () async {
      final chain = await _seedFullChain(db);
      final lease = await (db.select(
        db.leases,
      )..where((t) => t.id.equals(chain.leaseId))).getSingle();
      expect(lease.unitId, chain.unitId);
      expect(lease.tenantId, chain.tenantId);
      expect(lease.monthlyRent, 15000);
      expect(lease.active, isTrue);
    });

    test('deleting a unit cascades to its leases', () async {
      final chain = await _seedFullChain(db);
      await (db.delete(db.units)..where((t) => t.id.equals(chain.unitId))).go();

      final remainingLeases = await db.select(db.leases).get();
      expect(remainingLeases, isEmpty);
    });

    test('deleting a tenant cascades to its leases', () async {
      final chain = await _seedFullChain(db);
      await (db.delete(
        db.tenants,
      )..where((t) => t.id.equals(chain.tenantId))).go();

      final remainingLeases = await db.select(db.leases).get();
      expect(remainingLeases, isEmpty);
    });

    test('can end a lease (set inactive with an end date)', () async {
      final chain = await _seedFullChain(db);
      final lease = await (db.select(
        db.leases,
      )..where((t) => t.id.equals(chain.leaseId))).getSingle();

      await db
          .update(db.leases)
          .replace(
            lease.copyWith(
              active: false,
              endDate: Value(DateTime(2026, 6, 30)),
            ),
          );

      final updated = await (db.select(
        db.leases,
      )..where((t) => t.id.equals(chain.leaseId))).getSingle();
      expect(updated.active, isFalse);
      expect(updated.endDate, DateTime(2026, 6, 30));
    });
  });

  group('Payments', () {
    test('payment -> lease foreign key round-trip', () async {
      final chain = await _seedFullChain(db);
      final payment = await (db.select(
        db.payments,
      )..where((t) => t.id.equals(chain.paymentId))).getSingle();
      expect(payment.leaseId, chain.leaseId);
      expect(payment.month, 1);
      expect(payment.year, 2026);
      expect(payment.amountDue, 15000);
      expect(payment.amountPaid, 0); // default
      expect(payment.status, PaymentStatus.due);
    });

    test('deleting a lease cascades to its payments', () async {
      final chain = await _seedFullChain(db);
      await (db.delete(
        db.leases,
      )..where((t) => t.id.equals(chain.leaseId))).go();

      final remainingPayments = await db.select(db.payments).get();
      expect(remainingPayments, isEmpty);
    });

    test('recording a payment updates amountPaid and status', () async {
      final chain = await _seedFullChain(db);
      final payment = await (db.select(
        db.payments,
      )..where((t) => t.id.equals(chain.paymentId))).getSingle();

      await db
          .update(db.payments)
          .replace(
            payment.copyWith(
              amountPaid: 15000,
              status: PaymentStatus.paid,
              paymentMethod: const Value('bKash'),
              paymentDate: Value(DateTime(2026, 1, 5)),
            ),
          );

      final updated = await (db.select(
        db.payments,
      )..where((t) => t.id.equals(chain.paymentId))).getSingle();
      expect(updated.amountPaid, 15000);
      expect(updated.status, PaymentStatus.paid);
      expect(updated.paymentMethod, 'bKash');
      expect(updated.paymentDate, DateTime(2026, 1, 5));
    });

    test('a partial payment keeps amountDue and amountPaid distinct', () async {
      final chain = await _seedFullChain(db);
      final payment = await (db.select(
        db.payments,
      )..where((t) => t.id.equals(chain.paymentId))).getSingle();

      await db
          .update(db.payments)
          .replace(
            payment.copyWith(amountPaid: 5000, status: PaymentStatus.partial),
          );

      final updated = await (db.select(
        db.payments,
      )..where((t) => t.id.equals(chain.paymentId))).getSingle();
      expect(updated.amountDue, 15000);
      expect(updated.amountPaid, 5000);
      expect(updated.status, PaymentStatus.partial);
    });
  });

  group('Multi-level cascade', () {
    test(
      'deleting a property cascades all the way down to its payments',
      () async {
        final chain = await _seedFullChain(db);

        await (db.delete(
          db.properties,
        )..where((t) => t.id.equals(chain.propertyId))).go();

        expect(await db.select(db.buildings).get(), isEmpty);
        expect(await db.select(db.units).get(), isEmpty);
        expect(await db.select(db.leases).get(), isEmpty);
        expect(await db.select(db.payments).get(), isEmpty);
        // Tenants are NOT a child of property in the schema (they're
        // independent, only linked via lease) — must survive.
        expect(await db.select(db.tenants).get(), hasLength(1));
      },
    );

    test('deleting a building cascades to units, leases, and payments but not the property', () async {
      final chain = await _seedFullChain(db);

      await (db.delete(
        db.buildings,
      )..where((t) => t.id.equals(chain.buildingId))).go();

      expect(await db.select(db.units).get(), isEmpty);
      expect(await db.select(db.leases).get(), isEmpty);
      expect(await db.select(db.payments).get(), isEmpty);
      expect(await db.select(db.properties).get(), hasLength(1));
    });
  });
}
