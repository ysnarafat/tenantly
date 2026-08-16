import 'package:drift/drift.dart' show Value;
import 'package:drift/native.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tenantly_mobile/core/db/database.dart';
import 'package:tenantly_mobile/core/db/db_provider.dart';
import 'package:tenantly_mobile/core/db/tables.dart';
import 'package:tenantly_mobile/features/buildings/buildings_providers.dart';
import 'package:tenantly_mobile/features/leases/leases_providers.dart';
import 'package:tenantly_mobile/features/payments/payments_providers.dart';
import 'package:tenantly_mobile/features/properties/properties_providers.dart';
import 'package:tenantly_mobile/features/reports/reports_screen.dart';
import 'package:tenantly_mobile/features/units/units_providers.dart';

// ReportsScreen's stats providers are plain Providers that derive from five
// live stream providers via ref.watch. Pumping the widget straight away and
// hoping enough frames pass for all five to resolve is exactly the trap the
// non-widget reports_providers_test.dart hit first: on a fresh container,
// nothing has subscribed to those streams yet, so the first build sees them
// all still loading (-> zeros). Instead, build the ProviderContainer
// ourselves and await each stream's `.future` — which, being a first
// subscription, resolves with whatever's already in the DB — before the
// widget ever builds, then hand that pre-warmed container to the tree via
// UncontrolledProviderScope so the first frame already has real data.
Future<ProviderContainer> _settledContainer(AppDatabase db) async {
  final container = ProviderContainer(
    overrides: [appDatabaseProvider.overrideWithValue(db)],
  );
  await container.read(propertiesStreamProvider.future);
  await container.read(buildingsStreamProvider.future);
  await container.read(unitsStreamProvider.future);
  await container.read(leasesStreamProvider.future);
  await container.read(paymentsStreamProvider.future);
  return container;
}

Future<void> _pumpReportsScreen(
  WidgetTester tester,
  ProviderContainer container,
) async {
  // The screen's content is a single non-scrolled-in-test ListView; the
  // default 800x600 test surface only fits the Overview section, so the
  // Collection Summary and By Property sections below it are never built
  // (ListView only builds what's within the viewport) and any `find.text`
  // on them would report a false "not found". A tall viewport fits it all.
  tester.view.physicalSize = const Size(800, 3000);
  tester.view.devicePixelRatio = 1.0;
  addTearDown(tester.view.resetPhysicalSize);
  addTearDown(tester.view.resetDevicePixelRatio);

  await tester.pumpWidget(
    UncontrolledProviderScope(
      container: container,
      child: const MaterialApp(home: ReportsScreen()),
    ),
  );
  await tester.pump();
}

void main() {
  testWidgets(
    'shows a friendly empty state and non-crashing zero stats with no data',
    (tester) async {
      final db = AppDatabase.forTesting(NativeDatabase.memory());
      addTearDown(db.close);
      final container = await _settledContainer(db);
      addTearDown(container.dispose);

      await _pumpReportsScreen(tester, container);

      expect(find.text('No properties yet.'), findsOneWidget);
      // Occupancy and collection rate must render as 0%, not NaN% or a crash.
      expect(find.text('0%'), findsWidgets);

      await tester.runAsync(() async {
        await tester.pumpWidget(const SizedBox());
        await Future<void>.delayed(Duration.zero);
      });
    },
  );

  testWidgets('renders computed overview, collection, and breakdown stats', (
    tester,
  ) async {
    final db = AppDatabase.forTesting(NativeDatabase.memory());
    addTearDown(db.close);

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
    // Two units so occupancy comes out to a distinctive 50%, not 100%.
    final occupiedUnitId = await db
        .into(db.units)
        .insert(
          UnitsCompanion.insert(
            buildingId: buildingId,
            unitNumber: '101',
            unitType: UnitType.apartment,
          ),
        );
    await db
        .into(db.units)
        .insert(
          UnitsCompanion.insert(
            buildingId: buildingId,
            unitNumber: '102',
            unitType: UnitType.apartment,
          ),
        );
    final tenantId = await db
        .into(db.tenants)
        .insert(TenantsCompanion.insert(name: 'Jane Doe'));
    final leaseId = await db
        .into(db.leases)
        .insert(
          LeasesCompanion.insert(
            unitId: occupiedUnitId,
            tenantId: tenantId,
            monthlyRent: 15000,
            startDate: DateTime(2026, 1, 1),
          ),
        );
    await db
        .into(db.payments)
        .insert(
          PaymentsCompanion.insert(
            leaseId: leaseId,
            month: 1,
            year: 2026,
            amountDue: 15000,
            amountPaid: const Value(15000),
            status: PaymentStatus.paid,
          ),
        );

    final container = await _settledContainer(db);
    addTearDown(container.dispose);

    await _pumpReportsScreen(tester, container);

    // Overview.
    expect(find.text('50%'), findsOneWidget); // occupancy rate
    expect(find.text('৳15000'), findsWidgets); // rent roll + paid + revenue

    // Collection summary: fully paid -> 100% collection rate, 0 pending/overdue.
    expect(find.text('100%'), findsOneWidget); // collection rate
    expect(find.text('৳0'), findsWidgets); // pending + overdue

    // Property breakdown.
    expect(find.text('Sunset Apartments'), findsOneWidget);
    expect(find.text('1 building(s) · 1/2 units occupied'), findsOneWidget);

    await tester.runAsync(() async {
      await tester.pumpWidget(const SizedBox());
      await Future<void>.delayed(Duration.zero);
    });
  });
}
