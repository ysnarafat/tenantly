import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Polls [container]'s current value of [provider] until [predicate] returns
/// true on the resolved data, or throws after [timeout]. Used to assert a
/// live Drift stream provider has propagated a write without hardcoding a
/// fixed delay (Drift's watch() streams re-query asynchronously after any
/// write to a table they depend on, so a plain single `await` right after
/// the write is racy).
Future<T> waitForValue<T>(
  ProviderContainer container,
  ProviderListenable<AsyncValue<T>> provider,
  bool Function(T value) predicate, {
  Duration timeout = const Duration(seconds: 5),
}) async {
  final deadline = DateTime.now().add(timeout);
  while (DateTime.now().isBefore(deadline)) {
    final value = container.read(provider).valueOrNull;
    if (value != null && predicate(value)) {
      return value;
    }
    await Future<void>.delayed(const Duration(milliseconds: 10));
  }
  throw StateError(
    'waitForValue timed out after $timeout waiting for the predicate to hold. '
    'Last value: ${container.read(provider)}',
  );
}
