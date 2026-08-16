import 'package:drift/native.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tenantly_mobile/core/db/database.dart';
import 'package:tenantly_mobile/core/db/db_provider.dart';
import 'package:tenantly_mobile/core/db/tables.dart';
import 'package:tenantly_mobile/features/units/units_providers.dart';

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

  Future<int> seedBuilding({String name = 'Tower A'}) async {
    final propertyId = await db
        .into(db.properties)
        .insert(
          PropertiesCompanion.insert(
            name: 'Sunset Apartments',
            code: 'PROP-${name.hashCode}',
            propertyType: PropertyType.residential,
            address: '123 Main St',
          ),
        );
    return db
        .into(db.buildings)
        .insert(
          BuildingsCompanion.insert(
            propertyId: propertyId,
            name: name,
            code: 'BLD-${name.hashCode}',
            buildingType: BuildingType.residential,
          ),
        );
  }

  group('UnitsRepository', () {
    test('create inserts a row with a nullable floor', () async {
      final buildingId = await seedBuilding();
      final repo = UnitsRepository(db);
      final id = await repo.create(
        buildingId: buildingId,
        unitNumber: '101',
        unitType: UnitType.apartment,
      );

      final row = await (db.select(
        db.units,
      )..where((t) => t.id.equals(id))).getSingle();
      expect(row.buildingId, buildingId);
      expect(row.unitNumber, '101');
      expect(row.floor, isNull);
    });

    test('create stores an explicit floor', () async {
      final buildingId = await seedBuilding();
      final repo = UnitsRepository(db);
      final id = await repo.create(
        buildingId: buildingId,
        unitNumber: '304',
        unitType: UnitType.office,
        floor: 3,
      );

      final row = await (db.select(
        db.units,
      )..where((t) => t.id.equals(id))).getSingle();
      expect(row.floor, 3);
    });

    test('update and delete affect only the targeted row', () async {
      final buildingId = await seedBuilding();
      final repo = UnitsRepository(db);
      final id1 = await repo.create(
        buildingId: buildingId,
        unitNumber: '101',
        unitType: UnitType.apartment,
      );
      final id2 = await repo.create(
        buildingId: buildingId,
        unitNumber: '102',
        unitType: UnitType.shop,
      );

      final unit1 = await (db.select(
        db.units,
      )..where((t) => t.id.equals(id1))).getSingle();
      await repo.update(unit1.copyWith(unitNumber: '101-A'));
      await repo.delete(id2);

      final remaining = await db.select(db.units).get();
      expect(remaining, hasLength(1));
      expect(remaining.single.id, id1);
      expect(remaining.single.unitNumber, '101-A');
    });
  });

  group('unitsStreamProvider', () {
    test('joins each unit with its building\'s name', () async {
      final buildingId = await seedBuilding(name: 'Tower A');
      await container
          .read(unitsRepositoryProvider)
          .create(
            buildingId: buildingId,
            unitNumber: '101',
            unitType: UnitType.apartment,
          );

      final rows = await waitForValue(
        container,
        unitsStreamProvider,
        (rows) => rows.isNotEmpty,
      );
      expect(rows, hasLength(1));
      expect(rows.single.buildingName, 'Tower A');
      expect(rows.single.unit.unitNumber, '101');
    });

    test('orders units by unit number', () async {
      final buildingId = await seedBuilding();
      final repo = container.read(unitsRepositoryProvider);
      await repo.create(
        buildingId: buildingId,
        unitNumber: '201',
        unitType: UnitType.apartment,
      );
      await repo.create(
        buildingId: buildingId,
        unitNumber: '102',
        unitType: UnitType.apartment,
      );

      final rows = await waitForValue(
        container,
        unitsStreamProvider,
        (rows) => rows.length == 2,
      );
      expect(rows.map((r) => r.unit.unitNumber).toList(), ['102', '201']);
    });
  });

  group('unitByIdProvider', () {
    test('returns null for a nonexistent id', () async {
      final result = await container.read(unitByIdProvider(999).future);
      expect(result, isNull);
    });

    test('returns the row for an existing id', () async {
      final buildingId = await seedBuilding();
      final id = await container
          .read(unitsRepositoryProvider)
          .create(
            buildingId: buildingId,
            unitNumber: '101',
            unitType: UnitType.apartment,
            floor: 1,
          );

      final found = await container.read(unitByIdProvider(id).future);
      expect(found?.unitNumber, '101');
      expect(found?.floor, 1);
    });
  });
}
