import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/db/tables.dart';
import '../buildings/buildings_providers.dart';
import '../leases/leases_providers.dart';
import '../payments/payments_providers.dart';
import '../properties/properties_providers.dart';
import '../units/units_providers.dart';

class OverviewStats {
  OverviewStats({
    required this.propertyCount,
    required this.buildingCount,
    required this.unitCount,
    required this.activeLeaseCount,
    required this.occupancyRate,
    required this.totalRentRoll,
  });

  final int propertyCount;
  final int buildingCount;
  final int unitCount;
  final int activeLeaseCount;
  final double occupancyRate; // 0-100
  final double totalRentRoll;
}

class CollectionSummary {
  CollectionSummary({
    required this.totalDue,
    required this.totalPaid,
    required this.totalPending,
    required this.totalOverdue,
    required this.collectionRate,
  });

  final double totalDue;
  final double totalPaid;
  final double totalPending;
  final double totalOverdue;
  final double collectionRate; // 0-100
}

class PropertyBreakdownEntry {
  PropertyBreakdownEntry({
    required this.propertyName,
    required this.buildingCount,
    required this.unitCount,
    required this.occupiedUnitCount,
    required this.revenue,
  });

  final String propertyName;
  final int buildingCount;
  final int unitCount;
  final int occupiedUnitCount;
  final double revenue;
}

/// Every report/stat here is computed in Dart from the same live streams the
/// list screens already use — this dataset is small and entirely local, so a
/// handful of in-memory joins is simpler and easier to keep correct than
/// hand-written multi-table SQL, at this scale.
final overviewStatsProvider = Provider<OverviewStats>((ref) {
  final properties = ref.watch(propertiesStreamProvider).valueOrNull ?? [];
  final buildings = ref.watch(buildingsStreamProvider).valueOrNull ?? [];
  final units = ref.watch(unitsStreamProvider).valueOrNull ?? [];
  final leases = ref.watch(leasesStreamProvider).valueOrNull ?? [];

  final activeLeases = leases.where((l) => l.lease.active).toList();
  final occupancyRate = units.isEmpty
      ? 0.0
      : (activeLeases.length / units.length) * 100;
  final rentRoll = activeLeases.fold<double>(
    0,
    (sum, l) => sum + l.lease.monthlyRent,
  );

  return OverviewStats(
    propertyCount: properties.length,
    buildingCount: buildings.length,
    unitCount: units.length,
    activeLeaseCount: activeLeases.length,
    occupancyRate: occupancyRate,
    totalRentRoll: rentRoll,
  );
});

final collectionSummaryProvider = Provider<CollectionSummary>((ref) {
  final payments = ref.watch(paymentsStreamProvider).valueOrNull ?? [];

  var totalDue = 0.0;
  var totalPaid = 0.0;
  var totalOverdue = 0.0;
  for (final entry in payments) {
    final p = entry.payment;
    totalDue += p.amountDue;
    totalPaid += p.amountPaid;
    if (p.status == PaymentStatus.overdue) {
      totalOverdue += (p.amountDue - p.amountPaid);
    }
  }
  final totalPending = (totalDue - totalPaid).clamp(0, double.infinity);
  final collectionRate = totalDue == 0 ? 0.0 : (totalPaid / totalDue) * 100;

  return CollectionSummary(
    totalDue: totalDue,
    totalPaid: totalPaid,
    totalPending: totalPending.toDouble(),
    totalOverdue: totalOverdue,
    collectionRate: collectionRate,
  );
});

final propertyBreakdownProvider = Provider<List<PropertyBreakdownEntry>>((ref) {
  final properties = ref.watch(propertiesStreamProvider).valueOrNull ?? [];
  final buildings = ref.watch(buildingsStreamProvider).valueOrNull ?? [];
  final units = ref.watch(unitsStreamProvider).valueOrNull ?? [];
  final leases = ref.watch(leasesStreamProvider).valueOrNull ?? [];
  final payments = ref.watch(paymentsStreamProvider).valueOrNull ?? [];

  return properties.map((property) {
    final propertyBuildings = buildings
        .where((b) => b.building.propertyId == property.id)
        .toList();
    final buildingIds = propertyBuildings.map((b) => b.building.id).toSet();
    final propertyUnits = units
        .where((u) => buildingIds.contains(u.unit.buildingId))
        .toList();
    final unitIds = propertyUnits.map((u) => u.unit.id).toSet();
    final propertyLeases = leases
        .where((l) => unitIds.contains(l.lease.unitId))
        .toList();
    final occupiedUnitIds = propertyLeases
        .where((l) => l.lease.active)
        .map((l) => l.lease.unitId)
        .toSet();
    final leaseIds = propertyLeases.map((l) => l.lease.id).toSet();
    final revenue = payments
        .where((p) => leaseIds.contains(p.payment.leaseId))
        .fold<double>(0, (sum, p) => sum + p.payment.amountPaid);

    return PropertyBreakdownEntry(
      propertyName: property.name,
      buildingCount: propertyBuildings.length,
      unitCount: propertyUnits.length,
      occupiedUnitCount: occupiedUnitIds.length,
      revenue: revenue,
    );
  }).toList();
});
