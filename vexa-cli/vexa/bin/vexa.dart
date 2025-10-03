import 'dart:io';
import 'package:args/args.dart';
import 'package:vexa/vexa.dart' as vexa;

void main(List<String> arguments) async {
  final parser = ArgParser()
    ..addCommand('update', ArgParser()
      ..addCommand('start', ArgParser()
        ..addFlag('build-source', help: 'Build from source instead of using releases'))
      ..addCommand('status', ArgParser()
        ..addFlag('json', help: 'Output status as JSON'))
      ..addCommand('log'))
    ..addFlag('help', abbr: 'h', help: 'Show help information');

  final results = parser.parse(arguments);

  if (results['help'] as bool || arguments.isEmpty) {
    print('Vexa CLI - Update Helper');
    print('Version: 0.1.31');
    print('');
    print('Usage: vexa <command>');
    print('');
    print('Commands:');
    print('  update start [--build-source]  Start update process');
    print('  update status [--json]         Show update status');
    print('  update log                     Show update log');
    print('');
    print('Examples:');
    print('  vexa update start              Update using latest release');
    print('  vexa update start --build-source  Update by building from source');
    print('  vexa update status --json      Get update status as JSON');
    return;
  }

  try {
    final command = results.command;
    if (command?.name != 'update') {
      print('Unknown command: ${command?.name}');
      print('Use "vexa --help" for available commands');
      exit(1);
    }

    final subCommand = command?.command;
    switch (subCommand?.name) {
      case 'start':
        final buildSource = subCommand?['build-source'] as bool? ?? false;
        await vexa.startUpdate(buildSource);
        break;

      case 'status':
        final asJson = subCommand?['json'] as bool? ?? false;
        await vexa.showStatus(asJson);
        break;

      case 'log':
        await vexa.showLog();
        break;

      default:
        print('Unknown update command: ${subCommand?.name}');
        print('Use "vexa --help" for available commands');
        exit(1);
    }
  } catch (e) {
    print('Error: $e');
    exit(1);
  }
}