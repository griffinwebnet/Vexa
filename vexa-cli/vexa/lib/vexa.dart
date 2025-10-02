import 'dart:convert';
import 'dart:io';
import 'package:http/http.dart' as http;

/// Check for available updates from GitHub
Future<void> checkForUpdates() async {
  print('🔍 Checking for updates...');
  
  try {
    final response = await http.get(
      Uri.parse('https://api.github.com/repos/griffinwebnet/Vexa/releases'),
      headers: {'User-Agent': 'Vexa-CLI/0.0.4'},
    ).timeout(const Duration(seconds: 10));

    if (response.statusCode == 404) {
      print('ℹ️  No releases found (repository may not have releases yet)');
      return;
    }

    if (response.statusCode != 200) {
      print('❌ Failed to check for updates: HTTP ${response.statusCode}');
      return;
    }

    final releases = jsonDecode(response.body) as List;
    
    if (releases.isEmpty) {
      print('ℹ️  No releases available');
      return;
    }

    // Filter out pre-releases and get latest stable
    final stableReleases = releases.where((release) {
      final tagName = release['tag_name'] as String;
      return !tagName.contains('pre') && 
             !tagName.contains('alpha') && 
             !tagName.contains('beta');
    }).toList();

    if (stableReleases.isEmpty) {
      print('ℹ️  Only pre-releases available');
      return;
    }

    final latest = stableReleases.first;
    final latestVersion = latest['tag_name'] as String;
    final currentVersion = '0.0.4';

    print('📋 Current version: $currentVersion');
    print('📋 Latest version: $latestVersion');

    if (isUpdateAvailable(latestVersion, currentVersion)) {
      print('✅ Update available! Run "vexa upgrade" to install.');
      print('📝 Release notes: ${latest['body'] ?? 'No release notes'}');
    } else {
      print('✅ You are up to date!');
    }
  } catch (e) {
    print('❌ Failed to check for updates: $e');
  }
}

/// Perform system upgrade
Future<void> performUpgrade() async {
  print('🚀 Starting upgrade process...');
  print('⚠️  This may take up to 15 minutes and services will be unavailable during the upgrade.');
  print('');

  try {
    // Step 1: Update system packages
    print('📦 Updating base system packages...');
    await runCommand('sudo', ['apt', 'update'], showOutput: true);
    await runCommand('sudo', ['apt', 'upgrade', '-y'], showOutput: true);
    print('✅ Base system packages updated');
    print('');

    // Step 2: Update core system dependencies
    print('🔧 Updating core system dependencies...');
    await runCommand('sudo', ['apt', 'install', '-y', 'samba', 'bind9', 'krb5-user'], showOutput: true);
    print('✅ Core system dependencies updated');
    print('');

    // Step 3: Update main application
    print('📥 Updating main application...');
    await runCommand('git', ['pull', 'origin', 'main'], showOutput: true);
    print('✅ Main application updated');
    print('');

    // Step 4: Rebuild Go API
    print('🔨 Rebuilding API...');
    await runCommand('go', ['build'], workingDirectory: 'api', showOutput: true);
    print('✅ API rebuilt');
    print('');

    // Step 5: Rebuild React app
    print('🔨 Rebuilding web interface...');
    await runCommand('npm', ['run', 'build'], workingDirectory: 'web', showOutput: true);
    print('✅ Web interface rebuilt');
    print('');

    // Step 6: Restart services
    print('🔄 Restarting services...');
    await restartServices();
    print('');

    print('✅ Upgrade complete!');
    print('🎉 Vexa has been successfully updated and all services have been restarted.');
    print('🌐 You can now access the web interface at http://localhost:5173');
  } catch (e) {
    print('❌ Upgrade failed: $e');
    print('🔄 Attempting to restart services...');
    try {
      await restartServices();
    } catch (restartError) {
      print('❌ Failed to restart services: $restartError');
      print('⚠️  Please manually restart services: sudo systemctl restart vexa-api vexa-web');
    }
    rethrow;
  }
}

/// Show system status
Future<void> showStatus() async {
  print('📊 Vexa System Status');
  print('====================');
  print('');

  // Check API service
  try {
    final response = await http.get(
      Uri.parse('http://localhost:8080/health'),
    ).timeout(const Duration(seconds: 5));
    
    if (response.statusCode == 200) {
      final health = jsonDecode(response.body);
      print('✅ API Service: Running (${health['version']})');
    } else {
      print('❌ API Service: Not responding (HTTP ${response.statusCode})');
    }
  } catch (e) {
    print('❌ API Service: Not running ($e)');
  }

  // Check web service
  try {
    final response = await http.get(
      Uri.parse('http://localhost:5173'),
    ).timeout(const Duration(seconds: 5));
    
    if (response.statusCode == 200) {
      print('✅ Web Service: Running');
    } else {
      print('❌ Web Service: Not responding (HTTP ${response.statusCode})');
    }
  } catch (e) {
    print('❌ Web Service: Not running ($e)');
  }

  // Check system services
  try {
    final result = await Process.run('systemctl', ['is-active', 'samba-ad-dc']);
    if (result.exitCode == 0) {
      print('✅ Samba AD DC: ${result.stdout.toString().trim()}');
    } else {
      print('❌ Samba AD DC: Not running');
    }
  } catch (e) {
    print('❌ Samba AD DC: Status unknown');
  }

  try {
    final result = await Process.run('systemctl', ['is-active', 'bind9']);
    if (result.exitCode == 0) {
      print('✅ BIND9 DNS: ${result.stdout.toString().trim()}');
    } else {
      print('❌ BIND9 DNS: Not running');
    }
  } catch (e) {
    print('❌ BIND9 DNS: Status unknown');
  }

  print('');
  print('🌐 Web Interface: http://localhost:5173');
  print('🔌 API Endpoint: http://localhost:8080');
}

/// Restart all Vexa services
Future<void> restartServices() async {
  print('🔄 Restarting Vexa services...');
  
  final services = [
    'samba-ad-dc',
    'bind9',
    'vexa-api',
    'vexa-web',
  ];

  for (final service in services) {
    try {
      print('🔄 Restarting $service...');
      await runCommand('sudo', ['systemctl', 'restart', service]);
      print('✅ $service restarted');
    } catch (e) {
      print('⚠️  Failed to restart $service: $e');
    }
  }

  print('✅ All services restarted');
}

/// Run a command and optionally show output
Future<ProcessResult> runCommand(
  String command,
  List<String> arguments, {
  String? workingDirectory,
  bool showOutput = false,
}) async {
  final result = await Process.run(
    command,
    arguments,
    workingDirectory: workingDirectory,
  );

  if (showOutput && result.stdout.isNotEmpty) {
    print(result.stdout);
  }

  if (result.stderr.isNotEmpty) {
    print('Error: ${result.stderr}');
  }

  if (result.exitCode != 0) {
    throw Exception('Command failed: $command ${arguments.join(' ')}');
  }

  return result;
}

/// Check if an update is available
bool isUpdateAvailable(String latestVersion, String currentVersion) {
  // Remove 'v' prefix if present
  final latest = latestVersion.replaceFirst('v', '');
  final current = currentVersion.replaceFirst('v', '');

  // Simple version comparison (can be enhanced)
  return latest != current;
}