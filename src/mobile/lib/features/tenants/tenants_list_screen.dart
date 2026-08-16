import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'tenants_providers.dart';

class TenantsListScreen extends ConsumerWidget {
  const TenantsListScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final tenantsAsync = ref.watch(tenantsStreamProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Tenants')),
      body: tenantsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, stackTrace) =>
            Center(child: Text('Failed to load tenants: $error')),
        data: (tenants) {
          if (tenants.isEmpty) {
            return const Center(
              child: Text('No tenants yet. Tap + to add one.'),
            );
          }
          return ListView.builder(
            itemCount: tenants.length,
            itemBuilder: (context, index) {
              final tenant = tenants[index];
              return Dismissible(
                key: ValueKey(tenant.id),
                direction: DismissDirection.endToStart,
                background: Container(
                  color: Theme.of(context).colorScheme.errorContainer,
                  alignment: Alignment.centerRight,
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: const Icon(Icons.delete_outline),
                ),
                onDismissed: (_) {
                  ref.read(tenantsRepositoryProvider).delete(tenant.id);
                },
                child: ListTile(
                  title: Text(tenant.name),
                  subtitle: Text(tenant.phone ?? 'No phone on file'),
                  onTap: () => context.push('/tenants/${tenant.id}/edit'),
                ),
              );
            },
          );
        },
      ),
      floatingActionButton: FloatingActionButton(
        heroTag: 'tenants-fab',
        onPressed: () => context.push('/tenants/new'),
        tooltip: 'Add tenant',
        child: const Icon(Icons.add),
      ),
    );
  }
}
