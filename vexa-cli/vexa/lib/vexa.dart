import 'dart:convert';
import 'dart:io';

// Update status file
const statusFile = '/var/log/vexa/update_status.json';
const logFile = '/var/log/vexa/update.log';

/// Start the update process
Future<void> startUpdate(bool buildFromSource) async {
  // Create status directory if needed
  await Directory('/var/log/vexa').create(recursive: true);

  // Initialize status
  final status = {
    'status': 'starting',
    'progress': 0,
    'error': null,
    'completed': false,
  };
  await File(statusFile).writeAsString(jsonEncode(status));

  // Start update in background
  updateInBackground(buildFromSource);

  print('Update process started. Check status with "vexa update status"');
}

/// Show current update status
Future<void> showStatus(bool asJson) async {
  try {
    final status = await File(statusFile).readAsString();
    if (asJson) {
      print(status);
    } else {
      final data = jsonDecode(status) as Map<String, dynamic>;
      print('Status: ${data['status']}');
      print('Progress: ${data['progress']}%');
      if (data['error'] != null) {
        print('Error: ${data['error']}');
      }
      if (data['completed']) {
        print('Update completed successfully!');
      }
    }
  } catch (e) {
    if (asJson) {
      print(jsonEncode({
        'status': 'unknown',
        'progress': 0,
        'error': e.toString(),
        'completed': false,
      }));
    } else {
      print('No update status available');
    }
  }
}

/// Show update log
Future<void> showLog() async {
  try {
    final log = await File(logFile).readAsString();
    print(log);
  } catch (e) {
    print('No update log available');
  }
}

/// Update process that runs in background
Future<void> updateInBackground(bool buildFromSource) async {
  final log = File(logFile).openWrite();
  
  try {
    // Update status
    Future<void> updateStatus(String status, int progress) async {
      final data = {
        'status': status,
        'progress': progress,
        'error': null,
        'completed': false,
      };
      await File(statusFile).writeAsString(jsonEncode(data));
      log.writeln('[$status] Progress: $progress%');
    }

    // Step 1: Download source
    await updateStatus('downloading', 10);
    if (buildFromSource) {
      // Clone/pull repo
      await runCommand('git', ['clone', 'https://github.com/griffinwebnet/Vexa.git', '/opt/vexa/source']);
      await runCommand('git', ['checkout', 'main'], workingDirectory: '/opt/vexa/source');
    } else {
      // Download latest release
      // TODO: Implement release download
    }

    // Step 2: Install prerequisites
    await updateStatus('installing_prerequisites', 30);
    await runCommand('apt', ['update']);
    await runCommand('apt', ['install', '-y', 'golang', 'nodejs', 'npm', 'build-essential']);

    // Step 3: Build API
    await updateStatus('building_api', 50);
    await runCommand('go', ['build', '-o', '/usr/local/bin/vexa-api'], 
      workingDirectory: '/opt/vexa/source/api');

    // Step 4: Build Web UI
    await updateStatus('building_web', 70);
    await runCommand('npm', ['ci'], workingDirectory: '/opt/vexa/source/web');
    await runCommand('npm', ['run', 'build'], workingDirectory: '/opt/vexa/source/web');

    // Step 5: Install Web UI
    await updateStatus('installing', 90);
    await runCommand('rm', ['-rf', '/var/www/vexa']);
    await runCommand('mv', ['/opt/vexa/source/web/dist', '/var/www/vexa']);

    // Step 6: Restart services
    await updateStatus('restarting', 95);
    await runCommand('systemctl', ['restart', 'vexa-api']);
    await runCommand('systemctl', ['restart', 'nginx']);

    // Update completed
    final data = {
      'status': 'completed',
      'progress': 100,
      'error': null,
      'completed': true,
    };
    await File(statusFile).writeAsString(jsonEncode(data));
    log.writeln('[completed] Update finished successfully');

  } catch (e) {
    // Update failed
    final data = {
      'status': 'failed',
      'progress': 0,
      'error': e.toString(),
      'completed': false,
    };
    await File(statusFile).writeAsString(jsonEncode(data));
    log.writeln('[error] Update failed: $e');

    // Try to restart services
    try {
      await runCommand('systemctl', ['restart', 'vexa-api']);
      await runCommand('systemctl', ['restart', 'nginx']);
    } catch (e) {
      log.writeln('[error] Failed to restart services: $e');
    }
  } finally {
    await log.close();
  }
}

/// Run a command and log output
Future<void> runCommand(
  String command,
  List<String> arguments, {
  String? workingDirectory,
}) async {
  final result = await Process.run(
    command,
    arguments,
    workingDirectory: workingDirectory,
  );

  final log = File(logFile).openWrite(mode: FileMode.append);
  try {
    log.writeln('\$ $command ${arguments.join(' ')}');
    if (result.stdout.isNotEmpty) {
      log.writeln(result.stdout);
    }
    if (result.stderr.isNotEmpty) {
      log.writeln(result.stderr);
    }
  } finally {
    await log.close();
  }

  if (result.exitCode != 0) {
    throw Exception('Command failed: $command ${arguments.join(' ')}');
  }
}