import 'package:drift/native.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tenantly_mobile/core/db/database.dart';
import 'package:tenantly_mobile/core/db/db_provider.dart';
import 'package:tenantly_mobile/core/db/tables.dart';
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

  group('PropertiesRepository', () {
    test('create inserts a row and returns its id', () async {
      final repo = PropertiesRepository(db);
      final id = await repo.create(
        name: 'Sunset Apartments',
        code: 'PROP-001',
        propertyType: PropertyType.residential,
        address: '123 Main St',
      );

      final row = await (db.select(
        db.properties,
      )..where((t) => t.id.equals(id))).getSingle();
      expect(row.name, 'Sunset Apartments');
      expect(row.propertyType, PropertyType.residential);
    });

    test('update overwrites fields and bumps updatedAt', () async {
      final repo = PropertiesRepository(db);
      final id = await repo.create(
        name: 'Old Name',
        code: 'PROP-001',
        propertyType: PropertyType.residential,
        address: 'Old Address',
      );
      final original = await (db.select(
        db.properties,
      )..where((t) => t.id.equals(id))).getSingle();

      await repo.update(
        original.copyWith(name: 'New Name', propertyType: PropertyType.mixed),
      );

      final updated = await (db.select(
        db.properties,
      )..where((t) => t.id.equals(id))).getSingle();
      expect(updated.name, 'New Name');
      expect(updated.propertyType, PropertyType.mixed);
      expect(updated.code, 'PROP-001'); // untouched fields survive
      expect(updated.updatedAt.isAfter(original.updatedAt), isTrue);
    });

    test('delete removes the row', () async {
      final repo = PropertiesRepository(db);
      final id = await repo.create(
        name: 'Sunset Apartments',
        code: 'PROP-001',
        propertyType: PropertyType.residential,
        address: '123 Main St',
      );

      await repo.delete(id);

      expect(await db.select(db.properties).get(), isEmpty);
    });
  });

  group('propertiesStreamProvider', () {
    test('starts empty, then reflects an insert and a delete live', () async {
      expect(await container.read(propertiesStreamProvider.future), isEmpty);

      final repo = container.read(propertiesRepositoryProvider);
      final id = await repo.create(
        name: 'Sunset Apartments',
        code: 'PROP-001',
        propertyType: PropertyType.residential,
        address: '123 Main St',
      );

      final afterInsert = await waitForValue(
        container,
        propertiesStreamProvider,
        (rows) => rows.isNotEmpty,
      );
      expect(afterInsert, hasLength(1));
      expect(afterInsert.single.name, 'Sunset Apartments');

      await repo.delete(id);
      final afterDelete = await waitForValue(
        container,
        propertiesStreamProvider,
        (rows) => rows.isEmpty,
      );
      expect(afterDelete, isEmpty);
    });
  });

  group('propertyByIdProvider', () {
    test('returns null for a nonexistent id', () async {
      final result = await container.read(propertyByIdProvider(999).future);
      expect(result, isNull);
    });

    test('returns the row for an existing id', () async {
      final repo = container.read(propertiesRepositoryProvider);
      final id = await repo.create(
        name: 'Sunset Apartments',
        code: 'PROP-001',
        propertyType: PropertyType.residential,
        address: '123 Main St',
      );

      final found = await container.read(propertyByIdProvider(id).future);
      expect(found?.name, 'Sunset Apartments');
    });
  });
}
