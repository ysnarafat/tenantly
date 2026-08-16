import 'package:drift/native.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tenantly_mobile/core/db/database.dart';
import 'package:tenantly_mobile/core/db/db_provider.dart';
import 'package:tenantly_mobile/core/db/tables.dart';
import 'package:tenantly_mobile/features/leases/leases_providers.dart';

import '../../support/wait_for_value.dart';

void main() {
  late AppDatabase db;
  late ProviderContainer container;

  setUp(() {
    db = AppDatabase.forTesting(NativeDatabase.memory());
    container = ProviderContainer(
      overrides: [appDatabaseProvider.overrideWithValue(db)],
    );
  });

  tearDown(() async {
    container.dispose();
    await db.close();
  });

  Future<int> seedUnit({String unitNumber = '101'}) async {
    final propertyId = await db
        .into(db.properties)
        .insert(
          PropertiesCompanion.insert(
            name: 'Sunset Apartments',
            code: 'PROP-${unitNumber.hashCode}',
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
            code: 'BLD-${unitNumber.hashCode}',
            buildingType: BuildingType.residential,
          ),
        );
    return db
        .into(db.units)
        .insert(
          UnitsCompanion.insert(
            buildingId: buildingId,
            unitNumber: unitNumber,
            unitType: UnitType.apartment,
          ),
        );
  }

  Future<int> seedTenant({String name = 'Jane Doe'}) {
    return db.into(db.tenants).insert(TenantsCompanion.insert(name: name));
  }

  group('LeasesRepository', () {
    test('create inserts an active lease by default', () async {
      final unitId = await seedUnit();
      final tenantId = await seedTenant();
      final repo = LeasesRepository(db);
      final id = await repo.create(
        unitId: unitId,
        tenantId: tenantId,
        monthlyRent: 15000,
        startDate: DateTime(2026, 1, 1),
      );

      final row = await (db.select(
        db.leases,
      )..where((t) => t.id.equals(id))).getSingle();
      expect(row.unitId, unitId);
      expect(row.tenantId, tenantId);
      expect(row.monthlyRent, 15000);
      expect(row.active, isTrue);
      expect(row.endDate, isNull);
    });

    test('create can take an explicit end date and inactive flag', () async {
      final unitId = await seedUnit();
      final tenantId = await seedTenant();
      final repo = LeasesRepository(db);
      final id = await repo.create(
        unitId: unitId,
        tenantId: tenantId,
        monthlyRent: 15000,
        startDate: DateTime(2025, 1, 1),
        endDate: DateTime(2025, 12, 31),
        active: false,
      );

      final row = await (db.select(
        db.leases,
      )..where((t) => t.id.equals(id))).getSingle();
      expect(row.active, isFalse);
      expect(row.endDate, DateTime(2025, 12, 31));
    });

    test('update and delete affect only the targeted row', () async {
      final unitId = await seedUnit();
      final tenantId = await seedTenant();
      final repo = LeasesRepository(db);
      final id1 = await repo.create(
        unitId: unitId,
        tenantId: tenantId,
        monthlyRent: 15000,
        startDate: DateTime(2026, 1, 1),
      );
      final id2 = await repo.create(
        unitId: unitId,
        tenantId: tenantId,
        monthlyRent: 18000,
        startDate: DateTime(2026, 2, 1),
      );

      final lease1 = await (db.select(
        db.leases,
      )..where((t) => t.id.equals(id1))).getSingle();
      await repo.update(lease1.copyWith(monthlyRent: 16000));
      await repo.delete(id2);

      final remaining = await db.select(db.leases).get();
      expect(remaining, hasLength(1));
      expect(remaining.single.id, id1);
      expect(remaining.single.monthlyRent, 16000);
    });
  });

  group('leasesStreamProvider', () {
    test('joins each lease with its unit number and tenant name', () async {
      final unitId = await seedUnit(unitNumber: '101');
      final tenantId = await seedTenant(name: 'Jane Doe');
      await container
          .read(leasesRepositoryProvider)
          .create(
            unitId: unitId,
            tenantId: tenantId,
            monthlyRent: 15000,
            startDate: DateTime(2026, 1, 1),
          );

      final rows = await waitForValue(
        container,
        leasesStreamProvider,
        (rows) => rows.isNotEmpty,
      );
      expect(rows, hasLength(1));
      expect(rows.single.unitNumber, '101');
      expect(rows.single.tenantName, 'Jane Doe');
    });

    test('orders leases by most recent start date first', () async {
      final unitId = await seedUnit();
      final tenantId = await seedTenant();
      final repo = container.read(leasesRepositoryProvider);
      await repo.create(
        unitId: unitId,
        tenantId: tenantId,
        monthlyRent: 15000,
        startDate: DateTime(2024, 1, 1),
      );
      await repo.create(
        unitId: unitId,
        tenantId: tenantId,
        monthlyRent: 16000,
        startDate: DateTime(2026, 1, 1),
      );

      final rows = await waitForValue(
        container,
        leasesStreamProvider,
        (rows) => rows.length == 2,
      );
      expect(rows.map((r) => r.lease.startDate).toList(), [
        DateTime(2026, 1, 1),
        DateTime(2024, 1, 1),
      ]);
    });
  });

  group('leaseByIdProvider', () {
    test('returns null for a nonexistent id', () async {
      final result = await container.read(leaseByIdProvider(999).future);
      expect(result, isNull);
    });

    test('returns the row for an existing id', () async {
      final unitId = await seedUnit();
      final tenantId = await seedTenant();
      final id = await container
          .read(leasesRepositoryProvider)
          .create(
            unitId: unitId,
            tenantId: tenantId,
            monthlyRent: 15000,
            startDate: DateTime(2026, 1, 1),
          );

      final found = await container.read(leaseByIdProvider(id).future);
      expect(found?.monthlyRent, 15000);
    });
  });
}
