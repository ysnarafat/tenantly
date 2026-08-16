import 'package:drift/native.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tenantly_mobile/core/db/database.dart';
import 'package:tenantly_mobile/core/db/db_provider.dart';
import 'package:tenantly_mobile/core/db/tables.dart';
import 'package:tenantly_mobile/features/buildings/buildings_providers.dart';
import 'package:tenantly_mobile/features/properties/properties_providers.dart';

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

  Future<int> seedProperty({String name = 'Sunset Apartments'}) {
    return container
        .read(propertiesRepositoryProvider)
        .create(
          name: name,
          code: 'PROP-${name.hashCode}',
          propertyType: PropertyType.residential,
          address: '123 Main St',
        );
  }

  group('BuildingsRepository', () {
    test('create inserts a row scoped to its property', () async {
      final propertyId = await seedProperty();
      final repo = BuildingsRepository(db);
      final id = await repo.create(
        propertyId: propertyId,
        name: 'Tower A',
        code: 'BLD-001',
        buildingType: BuildingType.residential,
      );

      final row = await (db.select(
        db.buildings,
      )..where((t) => t.id.equals(id))).getSingle();
      expect(row.propertyId, propertyId);
      expect(row.name, 'Tower A');
    });

    test('update and delete affect only the targeted row', () async {
      final propertyId = await seedProperty();
      final repo = BuildingsRepository(db);
      final id1 = await repo.create(
        propertyId: propertyId,
        name: 'Tower A',
        code: 'BLD-001',
        buildingType: BuildingType.residential,
      );
      final id2 = await repo.create(
        propertyId: propertyId,
        name: 'Tower B',
        code: 'BLD-002',
        buildingType: BuildingType.commercial,
      );

      final building1 = await (db.select(
        db.buildings,
      )..where((t) => t.id.equals(id1))).getSingle();
      await repo.update(building1.copyWith(name: 'Tower A Renamed'));
      await repo.delete(id2);

      final remaining = await db.select(db.buildings).get();
      expect(remaining, hasLength(1));
      expect(remaining.single.id, id1);
      expect(remaining.single.name, 'Tower A Renamed');
    });
  });

  group('buildingsStreamProvider', () {
    test('joins each building with its property\'s name', () async {
      final propertyId = await seedProperty(name: 'Sunset Apartments');
      await container
          .read(buildingsRepositoryProvider)
          .create(
            propertyId: propertyId,
            name: 'Tower A',
            code: 'BLD-001',
            buildingType: BuildingType.residential,
          );

      final rows = await waitForValue(
        container,
        buildingsStreamProvider,
        (rows) => rows.isNotEmpty,
      );
      expect(rows, hasLength(1));
      expect(rows.single.propertyName, 'Sunset Apartments');
      expect(rows.single.building.name, 'Tower A');
    });

    test(
      'does not confuse buildings across two different properties',
      () async {
        final propertyA = await seedProperty(name: 'Property A');
        final propertyB = await seedProperty(name: 'Property B');
        final repo = container.read(buildingsRepositoryProvider);
        await repo.create(
          propertyId: propertyA,
          name: 'A Tower',
          code: 'BLD-A',
          buildingType: BuildingType.residential,
        );
        await repo.create(
          propertyId: propertyB,
          name: 'B Tower',
          code: 'BLD-B',
          buildingType: BuildingType.commercial,
        );

        final rows = await waitForValue(
          container,
          buildingsStreamProvider,
          (rows) => rows.length == 2,
        );
        final aEntry = rows.firstWhere((r) => r.building.name == 'A Tower');
        final bEntry = rows.firstWhere((r) => r.building.name == 'B Tower');
        expect(aEntry.propertyName, 'Property A');
        expect(bEntry.propertyName, 'Property B');
      },
    );
  });

  group('buildingByIdProvider', () {
    test('returns null for a nonexistent id', () async {
      final result = await container.read(buildingByIdProvider(999).future);
      expect(result, isNull);
    });
  });
}
