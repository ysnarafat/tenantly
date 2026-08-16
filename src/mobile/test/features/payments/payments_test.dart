import 'package:drift/native.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tenantly_mobile/core/db/database.dart';
import 'package:tenantly_mobile/core/db/db_provider.dart';
import 'package:tenantly_mobile/core/db/tables.dart';
import 'package:tenantly_mobile/features/payments/payments_providers.dart';

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

  Future<int> seedLease({
    String unitNumber = '101',
    String tenantName = 'Jane Doe',
  }) async {
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
    final unitId = await db
        .into(db.units)
        .insert(
          UnitsCompanion.insert(
            buildingId: buildingId,
            unitNumber: unitNumber,
            unitType: UnitType.apartment,
          ),
        );
    final tenantId = await db
        .into(db.tenants)
        .insert(TenantsCompanion.insert(name: tenantName));
    return db
        .into(db.leases)
        .insert(
          LeasesCompanion.insert(
            unitId: unitId,
            tenantId: tenantId,
            monthlyRent: 15000,
            startDate: DateTime(2026, 1, 1),
          ),
        );
  }

  group('PaymentsRepository', () {
    test('create defaults amountPaid to zero', () async {
      final leaseId = await seedLease();
      final repo = PaymentsRepository(db);
      final id = await repo.create(
        leaseId: leaseId,
        month: 1,
        year: 2026,
        amountDue: 15000,
        status: PaymentStatus.due,
      );

      final row = await (db.select(
        db.payments,
      )..where((t) => t.id.equals(id))).getSingle();
      expect(row.leaseId, leaseId);
      expect(row.amountDue, 15000);
      expect(row.amountPaid, 0);
      expect(row.status, PaymentStatus.due);
      expect(row.paymentMethod, isNull);
      expect(row.paymentDate, isNull);
    });

    test(
      'create stores an explicit paid amount, method, date, and notes',
      () async {
        final leaseId = await seedLease();
        final repo = PaymentsRepository(db);
        final id = await repo.create(
          leaseId: leaseId,
          month: 1,
          year: 2026,
          amountDue: 15000,
          amountPaid: 15000,
          status: PaymentStatus.paid,
          paymentMethod: 'bKash',
          paymentDate: DateTime(2026, 1, 5),
          notes: 'Paid in full',
        );

        final row = await (db.select(
          db.payments,
        )..where((t) => t.id.equals(id))).getSingle();
        expect(row.amountPaid, 15000);
        expect(row.status, PaymentStatus.paid);
        expect(row.paymentMethod, 'bKash');
        expect(row.paymentDate, DateTime(2026, 1, 5));
        expect(row.notes, 'Paid in full');
      },
    );

    test('update and delete affect only the targeted row', () async {
      final leaseId = await seedLease();
      final repo = PaymentsRepository(db);
      final id1 = await repo.create(
        leaseId: leaseId,
        month: 1,
        year: 2026,
        amountDue: 15000,
        status: PaymentStatus.due,
      );
      final id2 = await repo.create(
        leaseId: leaseId,
        month: 2,
        year: 2026,
        amountDue: 15000,
        status: PaymentStatus.due,
      );

      final payment1 = await (db.select(
        db.payments,
      )..where((t) => t.id.equals(id1))).getSingle();
      await repo.update(
        payment1.copyWith(amountPaid: 5000, status: PaymentStatus.partial),
      );
      await repo.delete(id2);

      final remaining = await db.select(db.payments).get();
      expect(remaining, hasLength(1));
      expect(remaining.single.id, id1);
      expect(remaining.single.amountPaid, 5000);
      expect(remaining.single.status, PaymentStatus.partial);
    });
  });

  group('paymentsStreamProvider', () {
    test(
      'joins each payment through its lease to the unit number and tenant name',
      () async {
        final leaseId = await seedLease(
          unitNumber: '101',
          tenantName: 'Jane Doe',
        );
        await container
            .read(paymentsRepositoryProvider)
            .create(
              leaseId: leaseId,
              month: 1,
              year: 2026,
              amountDue: 15000,
              status: PaymentStatus.due,
            );

        final rows = await waitForValue(
          container,
          paymentsStreamProvider,
          (rows) => rows.isNotEmpty,
        );
        expect(rows, hasLength(1));
        expect(rows.single.unitNumber, '101');
        expect(rows.single.tenantName, 'Jane Doe');
      },
    );

    test('orders payments by most recent year/month first', () async {
      final leaseId = await seedLease();
      final repo = container.read(paymentsRepositoryProvider);
      await repo.create(
        leaseId: leaseId,
        month: 1,
        year: 2026,
        amountDue: 15000,
        status: PaymentStatus.due,
      );
      await repo.create(
        leaseId: leaseId,
        month: 12,
        year: 2025,
        amountDue: 15000,
        status: PaymentStatus.due,
      );

      final rows = await waitForValue(
        container,
        paymentsStreamProvider,
        (rows) => rows.length == 2,
      );
      expect(rows.first.payment.year, 2026);
      expect(rows.first.payment.month, 1);
      expect(rows.last.payment.year, 2025);
    });
  });

  group('paymentByIdProvider', () {
    test('returns null for a nonexistent id', () async {
      final result = await container.read(paymentByIdProvider(999).future);
      expect(result, isNull);
    });

    test('returns the row for an existing id', () async {
      final leaseId = await seedLease();
      final id = await container
          .read(paymentsRepositoryProvider)
          .create(
            leaseId: leaseId,
            month: 1,
            year: 2026,
            amountDue: 15000,
            status: PaymentStatus.due,
          );

      final found = await container.read(paymentByIdProvider(id).future);
      expect(found?.amountDue, 15000);
    });
  });
}
