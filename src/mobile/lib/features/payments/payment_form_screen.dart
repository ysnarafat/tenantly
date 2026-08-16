import 'package:drift/drift.dart' show Value;
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/db/database.dart';
import '../../core/db/tables.dart';
import '../leases/leases_providers.dart';
import 'payments_providers.dart';

const _paymentMethods = ['Cash', 'Bank Transfer', 'bKash', 'Nagad', 'Cheque'];

class PaymentFormScreen extends ConsumerStatefulWidget {
  const PaymentFormScreen({super.key, this.paymentId});

  final int? paymentId;

  @override
  ConsumerState<PaymentFormScreen> createState() => _PaymentFormScreenState();
}

class _PaymentFormScreenState extends ConsumerState<PaymentFormScreen> {
  final _formKey = GlobalKey<FormState>();
  final _amountDueController = TextEditingController();
  final _amountPaidController = TextEditingController(text: '0');
  final _notesController = TextEditingController();
  int? _leaseId;
  int _month = DateTime.now().month;
  int _year = DateTime.now().year;
  PaymentStatus _status = PaymentStatus.due;
  String? _paymentMethod;
  DateTime? _paymentDate;
  bool _initialized = false;
  bool _saving = false;

  bool get _isEditMode => widget.paymentId != null;

  @override
  void dispose() {
    _amountDueController.dispose();
    _amountPaidController.dispose();
    _notesController.dispose();
    super.dispose();
  }

  void _populateFromExisting(PaymentRow payment) {
    if (_initialized) return;
    _leaseId = payment.leaseId;
    _month = payment.month;
    _year = payment.year;
    _amountDueController.text = payment.amountDue.toStringAsFixed(0);
    _amountPaidController.text = payment.amountPaid.toStringAsFixed(0);
    _status = payment.status;
    _paymentMethod = payment.paymentMethod;
    _paymentDate = payment.paymentDate;
    _notesController.text = payment.notes ?? '';
    _initialized = true;
  }

  Future<void> _pickPaymentDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _paymentDate ?? DateTime.now(),
      firstDate: DateTime(2000),
      lastDate: DateTime(2100),
    );
    if (picked != null) setState(() => _paymentDate = picked);
  }

  Future<void> _save() async {
    if (!_formKey.currentState!.validate()) return;
    if (_leaseId == null) return;
    setState(() => _saving = true);

    final amountDue = double.tryParse(_amountDueController.text.trim()) ?? 0;
    final amountPaid = double.tryParse(_amountPaidController.text.trim()) ?? 0;
    final notes = _notesController.text.trim();
    final repo = ref.read(paymentsRepositoryProvider);

    if (_isEditMode) {
      final existing = await ref.read(
        paymentByIdProvider(widget.paymentId!).future,
      );
      if (existing != null) {
        await repo.update(
          existing.copyWith(
            leaseId: _leaseId!,
            month: _month,
            year: _year,
            amountDue: amountDue,
            amountPaid: amountPaid,
            status: _status,
            paymentMethod: Value(_paymentMethod),
            paymentDate: Value(_paymentDate),
            notes: Value(notes.isEmpty ? null : notes),
          ),
        );
      }
    } else {
      await repo.create(
        leaseId: _leaseId!,
        month: _month,
        year: _year,
        amountDue: amountDue,
        amountPaid: amountPaid,
        status: _status,
        paymentMethod: _paymentMethod,
        paymentDate: _paymentDate,
        notes: notes.isEmpty ? null : notes,
      );
    }

    if (mounted) Navigator.of(context).pop();
  }

  Future<void> _delete() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete payment'),
        content: const Text('This cannot be undone.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.of(context).pop(true),
            child: const Text('Delete'),
          ),
        ],
      ),
    );
    if (confirmed == true) {
      await ref.read(paymentsRepositoryProvider).delete(widget.paymentId!);
      if (mounted) Navigator.of(context).pop();
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_isEditMode) {
      final paymentAsync = ref.watch(paymentByIdProvider(widget.paymentId!));
      return paymentAsync.when(
        loading: () =>
            const Scaffold(body: Center(child: CircularProgressIndicator())),
        error: (error, stackTrace) =>
            Scaffold(body: Center(child: Text('Failed to load: $error'))),
        data: (payment) {
          if (payment == null) {
            return const Scaffold(body: Center(child: Text('Not found')));
          }
          _populateFromExisting(payment);
          return _buildForm();
        },
      );
    }
    return _buildForm();
  }

  Widget _buildForm() {
    final leasesAsync = ref.watch(leasesStreamProvider);

    return Scaffold(
      appBar: AppBar(
        title: Text(_isEditMode ? 'Edit Payment' : 'Add Payment'),
        actions: [
          if (_isEditMode)
            IconButton(
              icon: const Icon(Icons.delete_outline),
              tooltip: 'Delete payment',
              onPressed: _delete,
            ),
        ],
      ),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            leasesAsync.when(
              loading: () => const LinearProgressIndicator(),
              error: (error, stackTrace) => const Text('Failed to load leases'),
              data: (leases) {
                if (leases.isEmpty) {
                  return const Text('Add a lease first before a payment.');
                }
                _leaseId ??= leases.first.lease.id;
                return DropdownButtonFormField<int>(
                  initialValue: _leaseId,
                  decoration: const InputDecoration(labelText: 'Lease'),
                  items: leases
                      .map(
                        (l) => DropdownMenuItem(
                          value: l.lease.id,
                          child: Text('${l.unitNumber} — ${l.tenantName}'),
                        ),
                      )
                      .toList(),
                  onChanged: (value) => setState(() => _leaseId = value),
                  validator: (value) => value == null ? 'Required' : null,
                );
              },
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  child: DropdownButtonFormField<int>(
                    initialValue: _month,
                    decoration: const InputDecoration(labelText: 'Month'),
                    items: List.generate(
                      12,
                      (i) => DropdownMenuItem(
                        value: i + 1,
                        child: Text(_monthLabel(i + 1)),
                      ),
                    ),
                    onChanged: (value) {
                      if (value != null) setState(() => _month = value);
                    },
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: DropdownButtonFormField<int>(
                    initialValue: _year,
                    decoration: const InputDecoration(labelText: 'Year'),
                    items: List.generate(5, (i) {
                      final year = DateTime.now().year - 2 + i;
                      return DropdownMenuItem(
                        value: year,
                        child: Text('$year'),
                      );
                    }),
                    onChanged: (value) {
                      if (value != null) setState(() => _year = value);
                    },
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _amountDueController,
              decoration: const InputDecoration(labelText: 'Amount due'),
              keyboardType: TextInputType.number,
              validator: (value) =>
                  (value == null || double.tryParse(value.trim()) == null)
                  ? 'Enter a valid amount'
                  : null,
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _amountPaidController,
              decoration: const InputDecoration(labelText: 'Amount paid'),
              keyboardType: TextInputType.number,
              validator: (value) =>
                  (value == null || double.tryParse(value.trim()) == null)
                  ? 'Enter a valid amount'
                  : null,
            ),
            const SizedBox(height: 12),
            DropdownButtonFormField<PaymentStatus>(
              initialValue: _status,
              decoration: const InputDecoration(labelText: 'Status'),
              items: PaymentStatus.values
                  .map((s) => DropdownMenuItem(value: s, child: Text(s.name)))
                  .toList(),
              onChanged: (value) {
                if (value != null) setState(() => _status = value);
              },
            ),
            const SizedBox(height: 12),
            DropdownButtonFormField<String>(
              initialValue: _paymentMethod,
              decoration: const InputDecoration(
                labelText: 'Payment method (optional)',
              ),
              items: _paymentMethods
                  .map((m) => DropdownMenuItem(value: m, child: Text(m)))
                  .toList(),
              onChanged: (value) => setState(() => _paymentMethod = value),
            ),
            const SizedBox(height: 12),
            ListTile(
              contentPadding: EdgeInsets.zero,
              title: const Text('Payment date (optional)'),
              subtitle: Text(
                _paymentDate == null ? 'Not set' : _formatDate(_paymentDate!),
              ),
              trailing: const Icon(Icons.calendar_today_outlined),
              onTap: _pickPaymentDate,
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _notesController,
              decoration: const InputDecoration(labelText: 'Notes (optional)'),
              maxLines: 2,
            ),
            const SizedBox(height: 24),
            FilledButton(
              onPressed: _saving ? null : _save,
              child: Text(_saving ? 'Saving...' : 'Save'),
            ),
          ],
        ),
      ),
    );
  }

  String _monthLabel(int month) => const [
    'January',
    'February',
    'March',
    'April',
    'May',
    'June',
    'July',
    'August',
    'September',
    'October',
    'November',
    'December',
  ][month - 1];

  String _formatDate(DateTime date) =>
      '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')}';
}
