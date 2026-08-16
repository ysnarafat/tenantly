import 'package:drift/drift.dart' show Value;
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/db/database.dart';
import 'tenants_providers.dart';

/// NID/phone are stored in plaintext locally for now — there is no
/// on-device encryption-at-rest infrastructure yet (see CONTRIBUTING.md and
/// the mobile setup plan). Not treated as solved; revisit before sync/export.
class TenantFormScreen extends ConsumerStatefulWidget {
  const TenantFormScreen({super.key, this.tenantId});

  final int? tenantId;

  @override
  ConsumerState<TenantFormScreen> createState() => _TenantFormScreenState();
}

class _TenantFormScreenState extends ConsumerState<TenantFormScreen> {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController();
  final _phoneController = TextEditingController();
  final _nidController = TextEditingController();
  bool _initialized = false;
  bool _saving = false;

  bool get _isEditMode => widget.tenantId != null;

  @override
  void dispose() {
    _nameController.dispose();
    _phoneController.dispose();
    _nidController.dispose();
    super.dispose();
  }

  void _populateFromExisting(TenantRow tenant) {
    if (_initialized) return;
    _nameController.text = tenant.name;
    _phoneController.text = tenant.phone ?? '';
    _nidController.text = tenant.nidNumber ?? '';
    _initialized = true;
  }

  Future<void> _save() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _saving = true);

    final phone = _phoneController.text.trim();
    final nid = _nidController.text.trim();
    final repo = ref.read(tenantsRepositoryProvider);
    if (_isEditMode) {
      final existing = await ref.read(
        tenantByIdProvider(widget.tenantId!).future,
      );
      if (existing != null) {
        await repo.update(
          existing.copyWith(
            name: _nameController.text.trim(),
            phone: Value(phone.isEmpty ? null : phone),
            nidNumber: Value(nid.isEmpty ? null : nid),
          ),
        );
      }
    } else {
      await repo.create(
        name: _nameController.text.trim(),
        phone: phone.isEmpty ? null : phone,
        nid: nid.isEmpty ? null : nid,
      );
    }

    if (mounted) Navigator.of(context).pop();
  }

  Future<void> _delete() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete tenant'),
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
      await ref.read(tenantsRepositoryProvider).delete(widget.tenantId!);
      if (mounted) Navigator.of(context).pop();
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_isEditMode) {
      final tenantAsync = ref.watch(tenantByIdProvider(widget.tenantId!));
      return tenantAsync.when(
        loading: () =>
            const Scaffold(body: Center(child: CircularProgressIndicator())),
        error: (error, stackTrace) =>
            Scaffold(body: Center(child: Text('Failed to load: $error'))),
        data: (tenant) {
          if (tenant == null) {
            return const Scaffold(body: Center(child: Text('Not found')));
          }
          _populateFromExisting(tenant);
          return _buildForm();
        },
      );
    }
    return _buildForm();
  }

  Widget _buildForm() {
    return Scaffold(
      appBar: AppBar(
        title: Text(_isEditMode ? 'Edit Tenant' : 'Add Tenant'),
        actions: [
          if (_isEditMode)
            IconButton(
              icon: const Icon(Icons.delete_outline),
              tooltip: 'Delete tenant',
              onPressed: _delete,
            ),
        ],
      ),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            TextFormField(
              controller: _nameController,
              decoration: const InputDecoration(labelText: 'Full name'),
              validator: (value) =>
                  (value == null || value.trim().isEmpty) ? 'Required' : null,
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _phoneController,
              decoration: const InputDecoration(labelText: 'Phone (optional)'),
              keyboardType: TextInputType.phone,
            ),
            const SizedBox(height: 12),
            TextFormField(
              controller: _nidController,
              decoration: const InputDecoration(labelText: 'NID (optional)'),
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
}
