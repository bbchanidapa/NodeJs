import { Injectable, Logger, OnModuleInit } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import * as admin from 'firebase-admin';

@Injectable()
export class FirebaseService implements OnModuleInit {
  private readonly logger = new Logger(FirebaseService.name);
  private app!: admin.app.App;

  constructor(private readonly config: ConfigService) {}

  onModuleInit() {
    const databaseURL =
      this.config.get<string>('FIREBASE_DATABASE_URL') ??
      this.databaseUrlFromFirebaseConfig();
    if (!databaseURL) {
      throw new Error(
        'FIREBASE_DATABASE_URL is not set. Copy .env.example to .env, or deploy on Firebase so FIREBASE_CONFIG includes databaseURL.',
      );
    }

    if (admin.apps.length > 0) {
      this.app = admin.app();
      return;
    }

    const credential = this.resolveCredential();

    this.app = admin.initializeApp({
      credential,
      databaseURL,
    });

    this.logger.log('Firebase Admin initialized');
  }

  private databaseUrlFromFirebaseConfig(): string | undefined {
    const raw = process.env.FIREBASE_CONFIG;
    if (!raw) {
      return undefined;
    }
    try {
      const cfg = JSON.parse(raw) as { databaseURL?: string };
      return cfg.databaseURL;
    } catch {
      return undefined;
    }
  }

  private resolveCredential(): admin.credential.Credential {
    const json = this.config.get<string>('FIREBASE_SERVICE_ACCOUNT_JSON');
    if (json) {
      return admin.credential.cert(JSON.parse(json) as admin.ServiceAccount);
    }

    const path =
      this.config.get<string>('GOOGLE_APPLICATION_CREDENTIALS') ??
      this.config.get<string>('FIREBASE_SERVICE_ACCOUNT_PATH');
    if (path) {
      return admin.credential.cert(path);
    }

    return admin.credential.applicationDefault();
  }

  getApp(): admin.app.App {
    return this.app;
  }

  getDatabase(): admin.database.Database {
    return admin.database(this.app);
  }
}
