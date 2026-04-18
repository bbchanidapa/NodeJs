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
      this.nonEmpty(this.config.get<string>('FIREBASE_DATABASE_URL')) ??
      this.databaseUrlFromFirebaseConfig() ??
      this.defaultRealtimeDatabaseUrlFromProjectId();
    if (!databaseURL) {
      throw new Error(
        'Set FIREBASE_DATABASE_URL in .env (see .env.example), or FIREBASE_PROJECT_ID for the default *-default-rtdb.firebaseio.com URL, or deploy with FIREBASE_CONFIG containing projectId.',
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

  private nonEmpty(value: string | undefined): string | undefined {
    return value && value.trim().length > 0 ? value.trim() : undefined;
  }

  private parsedFirebaseConfig():
    | { projectId?: string; databaseURL?: string }
    | undefined {
    const raw = process.env.FIREBASE_CONFIG;
    if (!raw) {
      return undefined;
    }
    try {
      return JSON.parse(raw) as { projectId?: string; databaseURL?: string };
    } catch {
      return undefined;
    }
  }

  private databaseUrlFromFirebaseConfig(): string | undefined {
    return this.parsedFirebaseConfig()?.databaseURL;
  }

  /** Default RTDB hostname; override with FIREBASE_DATABASE_URL if you use a regional *.firebasedatabase.app URL. */
  private defaultRealtimeDatabaseUrlFromProjectId(): string | undefined {
    const projectId =
      this.nonEmpty(this.config.get<string>('FIREBASE_PROJECT_ID')) ??
      this.nonEmpty(this.parsedFirebaseConfig()?.projectId);
    if (!projectId) {
      return undefined;
    }
    return `https://${projectId}-default-rtdb.firebaseio.com`;
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
