import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'reports_providers.dart';

class ReportsScreen extends ConsumerWidget {
  const ReportsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final overview = ref.watch(overviewStatsProvider);
    final collection = ref.watch(collectionSummaryProvider);
    final breakdown = ref.watch(propertyBreakdownProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Reports & Analysis')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          _SectionTitle('Overview'),
          _StatGrid(
            stats: [
              _Stat('Properties', '${overview.propertyCount}'),
              _Stat('Buildings', '${overview.buildingCount}'),
              _Stat('Units', '${overview.unitCount}'),
              _Stat('Active Leases', '${overview.activeLeaseCount}'),
              _Stat(
                'Occupancy',
                '${overview.occupancyRate.toStringAsFixed(0)}%',
              ),
              _Stat(
                'Monthly Rent Roll',
                '৳${overview.totalRentRoll.toStringAsFixed(0)}',
              ),
            ],
          ),
          const SizedBox(height: 24),
          _SectionTitle('Collection Summary'),
          _StatGrid(
            stats: [
              _Stat('Total Due', '৳${collection.totalDue.toStringAsFixed(0)}'),
              _Stat(
                'Total Collected',
                '৳${collection.totalPaid.toStringAsFixed(0)}',
              ),
              _Stat(
                'Pending',
                '৳${collection.totalPending.toStringAsFixed(0)}',
              ),
              _Stat(
                'Overdue',
                '৳${collection.totalOverdue.toStringAsFixed(0)}',
              ),
              _Stat(
                'Collection Rate',
                '${collection.collectionRate.toStringAsFixed(0)}%',
              ),
            ],
          ),
          const SizedBox(height: 24),
          _SectionTitle('By Property'),
          if (breakdown.isEmpty)
            const Padding(
              padding: EdgeInsets.symmetric(vertical: 16),
              child: Text('No properties yet.'),
            )
          else
            ...breakdown.map(
              (entry) => Card(
                margin: const EdgeInsets.only(bottom: 8),
                child: ListTile(
                  title: Text(entry.propertyName),
                  subtitle: Text(
                    '${entry.buildingCount} building(s) · '
                    '${entry.occupiedUnitCount}/${entry.unitCount} units occupied',
                  ),
                  trailing: Text(
                    '৳${entry.revenue.toStringAsFixed(0)}',
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                ),
              ),
            ),
        ],
      ),
    );
  }
}

class _SectionTitle extends StatelessWidget {
  const _SectionTitle(this.title);

  final String title;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Text(title, style: Theme.of(context).textTheme.titleMedium),
    );
  }
}

class _Stat {
  _Stat(this.label, this.value);

  final String label;
  final String value;
}

class _StatGrid extends StatelessWidget {
  const _StatGrid({required this.stats});

  final List<_Stat> stats;

  @override
  Widget build(BuildContext context) {
    return GridView.count(
      crossAxisCount: 2,
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      childAspectRatio: 2.2,
      crossAxisSpacing: 8,
      mainAxisSpacing: 8,
      children: stats
          .map(
            (stat) => Card(
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(
                      stat.value,
                      style: Theme.of(context).textTheme.headlineSmall,
                    ),
                    Text(
                      stat.label,
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ],
                ),
              ),
            ),
          )
          .toList(),
    );
  }
}
