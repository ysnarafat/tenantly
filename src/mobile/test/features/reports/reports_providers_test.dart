import 'package:drift/native.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tenantly_mobile/core/db/database.dart';
import 'package:tenantly_mobile/core/db/db_provider.dart';
import 'package:tenantly_mobile/core/db/tables.dart';
import 'package:tenantly_mobile/features/buildings/buildings_providers.dart';
import 'package:tenantly_mobile/features/leases/leases_providers.dart';
import 'package:tenantly_mobile/features/payments/payments_providers.dart';
import 'package:tenantly_mobile/features/properties/properties_providers.dart';
import 'package:tenantly_mobile/features/reports/reports_providers.dart';
import 'package:tenantly_mobile/features/tenants/tenants_providers.dart';
import 'package:tenantly_mobile/features/units/units_providers.dart';

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

  // Seeds property -> building -> unit -> tenant -> lease, and returns every
  // id, so each test only has to specify the fields it actually cares about.
  Future<
    ({int propertyId, int buildingId, int unitId, int tenantId, int leaseId})
  >
  seedLease({
    String propertyName = 'Sunset Apartments',
    String unitNumber = '101',
    double monthlyRent = 15000,
    bool active = true,
  }) async {
    final propertiesRepo = container.read(propertiesRepositoryProvider);
    final buildingsRepo = container.read(buildingsRepositoryProvider);
    final unitsRepo = container.read(unitsRepositoryProvider);
    final tenantsRepo = container.read(tenantsRepositoryProvider);
    final leasesRepo = container.read(leasesRepositoryProvider);

    final propertyId = await propertiesRepo.create(
      name: propertyName,
      code: 'PROP-${propertyName.hashCode}-${unitNumber.hashCode}',
      propertyType: PropertyType.residential,
      address: '123 Main St',
    );
    final buildingId = await buildingsRepo.create(
      propertyId: propertyId,
      name: 'Tower A',
      code: 'BLD-${unitNumber.hashCode}',
      buildingType: BuildingType.residential,
    );
    final unitId = await unitsRepo.create(
      buildingId: buildingId,
      unitNumber: unitNumber,
      unitType: UnitType.apartment,
    );
    final tenantId = await tenantsRepo.create(name: 'Tenant $unitNumber');
    final leaseId = await leasesRepo.create(
      unitId: unitId,
      tenantId: tenantId,
      monthlyRent: monthlyRent,
      startDate: DateTime(2026, 1, 1),
      active: active,
    );

    return (
      propertyId: propertyId,
      buildingId: buildingId,
      unitId: unitId,
      tenantId: tenantId,
      leaseId: leaseId,
    );
  }

  // The overview/collection/breakdown providers derive from these five
  // stream providers via ref.watch, but a plain Provider's build only runs
  // (subscribing to those streams for the first time) when it's first read.
  // Awaiting each stream's `.future` here — after all seeding writes have
  // completed — forces that first subscription (and its initial query) to
  // happen now, so the derived provider sees current data on its first read
  // rather than each dependency's still-loading initial state.
  Future<void> settle() async {
    await container.read(propertiesStreamProvider.future);
    await container.read(buildingsStreamProvider.future);
    await container.read(unitsStreamProvider.future);
    await container.read(leasesStreamProvider.future);
    await container.read(paymentsStreamProvider.future);
  }

  Future<void> seedPayment({
    required int leaseId,
    required double amountDue,
    required double amountPaid,
    required PaymentStatus status,
  }) async {
    await container
        .read(paymentsRepositoryProvider)
        .create(
          leaseId: leaseId,
          month: 1,
          year: 2026,
          amountDue: amountDue,
          amountPaid: amountPaid,
          status: status,
        );
  }

  group('overviewStatsProvider', () {
    test('is all zero on an empty database', () {
      final stats = container.read(overviewStatsProvider);
      expect(stats.propertyCount, 0);
      expect(stats.buildingCount, 0);
      expect(stats.unitCount, 0);
      expect(stats.activeLeaseCount, 0);
      expect(stats.occupancyRate, 0.0);
      expect(stats.totalRentRoll, 0.0);
    });

    test('counts each entity across multiple properties', () async {
      await seedLease(propertyName: 'Property A', unitNumber: '101');
      await seedLease(propertyName: 'Property B', unitNumber: '201');

      await settle();

      final stats = container.read(overviewStatsProvider);
      expect(stats.propertyCount, 2);
      expect(stats.buildingCount, 2);
      expect(stats.unitCount, 2);
      expect(stats.activeLeaseCount, 2);
    });

    test('excludes inactive leases from activeLeaseCount, occupancy, and rent roll', () async {
      await seedLease(unitNumber: '101', monthlyRent: 15000, active: true);
      await seedLease(unitNumber: '102', monthlyRent: 20000, active: false);

      await settle();

      final stats = container.read(overviewStatsProvider);
      expect(stats.unitCount, 2);
      expect(stats.activeLeaseCount, 1);
      expect(stats.occupancyRate, 50.0);
      expect(stats.totalRentRoll, 15000);
    });

    test(
      'occupancy rate is 0, not NaN, when there are units but no leases',
      () async {
        final buildingId = (await seedLease(unitNumber: '101')).buildingId;
        await container
            .read(unitsRepositoryProvider)
            .create(
              buildingId: buildingId,
              unitNumber: '999',
              unitType: UnitType.apartment,
            );

        await settle();

        final stats = container.read(overviewStatsProvider);
        expect(stats.unitCount, 2);
        expect(stats.activeLeaseCount, 1);
        expect(stats.occupancyRate, 50.0);
        expect(stats.occupancyRate.isNaN, isFalse);
      },
    );
  });

  group('collectionSummaryProvider', () {
    test('is all zero, with a 0 collection rate, on an empty database', () {
      final summary = container.read(collectionSummaryProvider);
      expect(summary.totalDue, 0.0);
      expect(summary.totalPaid, 0.0);
      expect(summary.totalPending, 0.0);
      expect(summary.totalOverdue, 0.0);
      expect(summary.collectionRate, 0.0);
      expect(summary.collectionRate.isNaN, isFalse);
    });

    test(
      'sums due/paid across multiple payments and computes collection rate',
      () async {
        final lease = await seedLease();
        await seedPayment(
          leaseId: lease.leaseId,
          amountDue: 15000,
          amountPaid: 15000,
          status: PaymentStatus.paid,
        );
        await seedPayment(
          leaseId: lease.leaseId,
          amountDue: 15000,
          amountPaid: 5000,
          status: PaymentStatus.partial,
        );

        await settle();

        final summary = container.read(collectionSummaryProvider);
        expect(summary.totalDue, 30000);
        expect(summary.totalPaid, 20000);
        expect(summary.totalPending, 10000);
        expect(summary.collectionRate, closeTo(66.67, 0.1));
      },
    );

    test(
      'only sums the unpaid gap of overdue payments toward totalOverdue',
      () async {
        final lease = await seedLease();
        await seedPayment(
          leaseId: lease.leaseId,
          amountDue: 15000,
          amountPaid: 0,
          status: PaymentStatus.overdue,
        );
        // A due (not yet overdue) unpaid payment must NOT count as overdue.
        await seedPayment(
          leaseId: lease.leaseId,
          amountDue: 15000,
          amountPaid: 0,
          status: PaymentStatus.due,
        );

        await settle();

        final summary = container.read(collectionSummaryProvider);
        expect(summary.totalOverdue, 15000);
        expect(summary.totalDue, 30000);
      },
    );

    test(
      'clamps totalPending at zero even if payments overpay in aggregate',
      () async {
        final leaseA = await seedLease(unitNumber: '101');
        final leaseB = await seedLease(unitNumber: '102');
        // Overpaid relative to amountDue.
        await seedPayment(
          leaseId: leaseA.leaseId,
          amountDue: 10000,
          amountPaid: 12000,
          status: PaymentStatus.paid,
        );
        await seedPayment(
          leaseId: leaseB.leaseId,
          amountDue: 5000,
          amountPaid: 1000,
          status: PaymentStatus.partial,
        );

        await settle();

        final summary = container.read(collectionSummaryProvider);
        // 15000 due, 13000 paid -> 2000 pending, never negative.
        expect(summary.totalPending, 2000);
      },
    );
  });

  group('propertyBreakdownProvider', () {
    test('is empty on an empty database', () {
      expect(container.read(propertyBreakdownProvider), isEmpty);
    });

    test(
      'isolates buildings, units, occupancy, and revenue per property',
      () async {
        final leaseA = await seedLease(
          propertyName: 'Property A',
          unitNumber: '101',
        );
        final leaseB = await seedLease(
          propertyName: 'Property B',
          unitNumber: '201',
        );
        await seedPayment(
          leaseId: leaseA.leaseId,
          amountDue: 15000,
          amountPaid: 15000,
          status: PaymentStatus.paid,
        );
        await seedPayment(
          leaseId: leaseB.leaseId,
          amountDue: 20000,
          amountPaid: 5000,
          status: PaymentStatus.partial,
        );

        await settle();

        final breakdown = container.read(propertyBreakdownProvider);
        expect(breakdown, hasLength(2));

        final entryA = breakdown.firstWhere(
          (e) => e.propertyName == 'Property A',
        );
        expect(entryA.buildingCount, 1);
        expect(entryA.unitCount, 1);
        expect(entryA.occupiedUnitCount, 1);
        expect(entryA.revenue, 15000);

        final entryB = breakdown.firstWhere(
          (e) => e.propertyName == 'Property B',
        );
        expect(entryB.buildingCount, 1);
        expect(entryB.unitCount, 1);
        expect(entryB.occupiedUnitCount, 1);
        expect(entryB.revenue, 5000);
      },
    );

    test(
      'a building with no units still reports zero counts and revenue',
      () async {
        final lease = await seedLease();
        await container
            .read(buildingsRepositoryProvider)
            .create(
              propertyId: lease.propertyId,
              name: 'Empty Tower',
              code: 'BLD-EMPTY',
              buildingType: BuildingType.commercial,
            );

        await settle();

        final breakdown = container.read(propertyBreakdownProvider);
        expect(breakdown, hasLength(1));
        expect(breakdown.single.buildingCount, 2);
        expect(breakdown.single.unitCount, 1);
      },
    );

    test('an inactive lease\'s unit does not count as occupied', () async {
      await seedLease(unitNumber: '101', active: false);

      await settle();

      final breakdown = container.read(propertyBreakdownProvider);
      expect(breakdown.single.unitCount, 1);
      expect(breakdown.single.occupiedUnitCount, 0);
    });
  });
}
