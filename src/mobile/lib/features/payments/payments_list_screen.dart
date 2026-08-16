import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/db/tables.dart';
import 'payments_providers.dart';

const _monthNames = [
  'Jan',
  'Feb',
  'Mar',
  'Apr',
  'May',
  'Jun',
  'Jul',
  'Aug',
  'Sep',
  'Oct',
  'Nov',
  'Dec',
];

class PaymentsListScreen extends ConsumerWidget {
  const PaymentsListScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final paymentsAsync = ref.watch(paymentsStreamProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Payments')),
      body: paymentsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, stackTrace) =>
            Center(child: Text('Failed to load payments: $error')),
        data: (payments) {
          if (payments.isEmpty) {
            return const Center(
              child: Text('No payments yet. Tap + to add one.'),
            );
          }
          return ListView.builder(
            itemCount: payments.length,
            itemBuilder: (context, index) {
              final entry = payments[index];
              final payment = entry.payment;
              return Dismissible(
                key: ValueKey(payment.id),
                direction: DismissDirection.endToStart,
                background: Container(
                  color: Theme.of(context).colorScheme.errorContainer,
                  alignment: Alignment.centerRight,
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  child: const Icon(Icons.delete_outline),
                ),
                onDismissed: (_) {
                  ref.read(paymentsRepositoryProvider).delete(payment.id);
                },
                child: ListTile(
                  title: Text(
                    '${entry.unitNumber} — ${entry.tenantName}  '
                    '(${_monthNames[payment.month - 1]} ${payment.year})',
                  ),
                  subtitle: Text(
                    '৳${payment.amountPaid.toStringAsFixed(0)} / '
                    '৳${payment.amountDue.toStringAsFixed(0)}',
                  ),
                  trailing: _StatusChip(status: payment.status),
                  onTap: () => context.push('/payments/${payment.id}/edit'),
                ),
              );
            },
          );
        },
      ),
      floatingActionButton: FloatingActionButton(
        heroTag: 'payments-fab',
        onPressed: () => context.push('/payments/new'),
        tooltip: 'Add payment',
        child: const Icon(Icons.add),
      ),
    );
  }
}

class _StatusChip extends StatelessWidget {
  const _StatusChip({required this.status});

  final PaymentStatus status;

  @override
  Widget build(BuildContext context) {
    final (label, color) = switch (status) {
      PaymentStatus.paid => ('Paid', Colors.green),
      PaymentStatus.partial => ('Partial', Colors.orange),
      PaymentStatus.due => ('Due', Colors.blueGrey),
      PaymentStatus.overdue => ('Overdue', Colors.red),
    };
    return Chip(
      label: Text(label, style: const TextStyle(fontSize: 11)),
      backgroundColor: color.withValues(alpha: 0.15),
      labelStyle: TextStyle(color: color),
      padding: EdgeInsets.zero,
      visualDensity: VisualDensity.compact,
    );
  }
}
