import 'dart:math';
import 'package:uuid/uuid.dart';

class Device {
  final String oauthId;
  final Map<String, String> initialHeaders;
  final Map<String, String> headers;
  final String userAgent;

  Device._({
    required this.oauthId,
    required this.initialHeaders,
    required this.headers,
    required this.userAgent,
  });

  static Device android() {
    final uuid = Uuid().v4();
    const androidAppVersion = '2024.15.0'; // A recent version
    final androidVersion = (Random().nextInt(6) + 9).toString(); // Android 9 to 14

    final userAgent = 'Reddit/$androidAppVersion/Android $androidVersion';

    final qos = (Random().nextDouble() * 99 + 1).toStringAsFixed(3);

    // This is a simplified version of the tegen generator in redlib
    final codecs = 'available-codecs=video/avc, video/hevc, video/x-vnd.on2.vp9';

    final headers = <String, String>{
      'User-Agent': userAgent,
      'x-reddit-retry': 'algo=no-retries',
      'x-reddit-compression': '1',
      'x-reddit-qos': qos,
      'x-reddit-media-codecs': codecs,
      'Content-Type': 'application/json; charset=UTF-8',
      'client-vendor-id': uuid,
      'X-Reddit-Device-Id': uuid,
    };

    return Device._(
      oauthId: 'ohXpoqrZYub1kg',
      initialHeaders: headers,
      headers: headers,
      userAgent: userAgent,
    );
  }
}
