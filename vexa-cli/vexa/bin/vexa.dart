import 'dart:io';
import 'package:args/args.dart';
import 'package:vexa/vexa.dart' as vexa;

void main(List<String> arguments) async {
  final parser = ArgParser()
    ..addCommand('update')
    ..addCommand('upgrade')
    ..addCommand('status')
    ..addCommand('restart')
    ..addFlag('help', abbr: 'h', help: 'Show help information');

  final results = parser.parse(arguments);

  if (results['help'] as bool || arguments.isEmpty) {
    print('Vexa CLI - Directory Services Management Tool');
    print('Version: 0.0.2-prealpha');
    print('');
    print('Usage: vexa <command>');
    print('');
    print('Commands:');
    print('  update    Check for available updates');
    print('  upgrade   Install updates and restart services');
    print('  status    Show system status');
    print('  restart   Restart all Vexa services');
    print('');
    print('Examples:');
    print('  vexa update');
    print('  vexa upgrade');
    print('  vexa status');
    return;
  }

  try {
    switch (results.command?.name) {
      case 'update':
        await vexa.checkForUpdates();
        break;
      case 'upgrade':
        await vexa.performUpgrade();
        break;
      case 'status':
        await vexa.showStatus();
        break;
      case 'restart':
        await vexa.restartServices();
        break;
      default:
        print('Unknown command: ${results.command?.name}');
        print('Use "vexa --help" for available commands');
        exit(1);
    }
  } catch (e) {
    print('Error: $e');
    exit(1);
  }
}