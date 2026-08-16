import 'package:drift/native.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tenantly_mobile/core/db/database.dart';
import 'package:tenantly_mobile/core/db/db_provider.dart';
import 'package:tenantly_mobile/features/tenants/tenants_providers.dart';

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

  group('TenantsRepository', () {
    test('create inserts a row with optional phone and NID', () async {
      final repo = TenantsRepository(db);
      final id = await repo.create(
        name: 'Jane Doe',
        phone: '01712345678',
        nid: '1234567890',
      );

      final row = await (db.select(
        db.tenants,
      )..where((t) => t.id.equals(id))).getSingle();
      expect(row.name, 'Jane Doe');
      expect(row.phone, '01712345678');
      expect(row.nidNumber, '1234567890');
    });

    test('create allows omitting phone and NID entirely', () async {
      final repo = TenantsRepository(db);
      final id = await repo.create(name: 'No Contact Info');

      final row = await (db.select(
        db.tenants,
      )..where((t) => t.id.equals(id))).getSingle();
      expect(row.phone, isNull);
      expect(row.nidNumber, isNull);
    });

    test('update and delete affect only the targeted row', () async {
      final repo = TenantsRepository(db);
      final id1 = await repo.create(name: 'Jane Doe');
      final id2 = await repo.create(name: 'John Smith');

      final tenant1 = await (db.select(
        db.tenants,
      )..where((t) => t.id.equals(id1))).getSingle();
      await repo.update(tenant1.copyWith(name: 'Jane Renamed'));
      await repo.delete(id2);

      final remaining = await db.select(db.tenants).get();
      expect(remaining, hasLength(1));
      expect(remaining.single.id, id1);
      expect(remaining.single.name, 'Jane Renamed');
    });
  });

  group('tenantsStreamProvider', () {
    test('orders tenants alphabetically by name', () async {
      final repo = container.read(tenantsRepositoryProvider);
      await repo.create(name: 'Zed Tenant');
      await repo.create(name: 'Anna Tenant');

      final rows = await waitForValue(
        container,
        tenantsStreamProvider,
        (rows) => rows.length == 2,
      );
      expect(rows.map((r) => r.name).toList(), ['Anna Tenant', 'Zed Tenant']);
    });
  });

  group('tenantByIdProvider', () {
    test('returns null for a nonexistent id', () async {
      final result = await container.read(tenantByIdProvider(999).future);
      expect(result, isNull);
    });

    test('returns the row for an existing id', () async {
      final id = await container
          .read(tenantsRepositoryProvider)
          .create(name: 'Jane Doe', phone: '01712345678');

      final found = await container.read(tenantByIdProvider(id).future);
      expect(found?.name, 'Jane Doe');
      expect(found?.phone, '01712345678');
    });
  });
}
