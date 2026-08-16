import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'leases_providers.dart';

class LeasesListScreen extends ConsumerWidget {
  const LeasesListScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final leasesAsync = ref.watch(leasesStreamProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Leases')),
      body: leasesAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, stackTrace) =>
            Center(child: Text('Failed to load leases: $error')),
        data: (leases) {
          if (leases.isEmpty) {
            return const Center(
              child: Text('No leases yet. Tap + to add one.'),
            );
          }
          return ListView.builder(
            itemCount: leases.length,
            itemBuilder: (context, index) {
              final entry = leases[index];
              final lease = entry.lease;
              return Dismissible(
                key: ValueKey(lease.id),
                direction: DismissDirection.endToStart,
                background: Container(
                  color: Theme.of(context).colorScheme.errorContainer,
                  alignment: Alignment.centerRight,
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: const Icon(Icons.delete_outline),
                ),
                onDismissed: (_) {
                  ref.read(leasesRepositoryProvider).delete(lease.id);
                },
                child: ListTile(
                  title: Text('${entry.unitNumber} — ${entry.tenantName}'),
                  subtitle: Text(
                    '৳${lease.monthlyRent.toStringAsFixed(0)}/mo · '
                    'from ${_formatDate(lease.startDate)}'
                    '${lease.active ? '' : ' · inactive'}',
                  ),
                  onTap: () => context.push('/leases/${lease.id}/edit'),
                ),
              );
            },
          );
        },
      ),
      floatingActionButton: FloatingActionButton(
        heroTag: 'leases-fab',
        onPressed: () => context.push('/leases/new'),
        tooltip: 'Add lease',
        child: const Icon(Icons.add),
      ),
    );
  }

  String _formatDate(DateTime date) =>
      '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')}';
}
