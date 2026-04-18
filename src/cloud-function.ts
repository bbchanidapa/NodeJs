import { NestFactory } from '@nestjs/core';
import { ExpressAdapter } from '@nestjs/platform-express';
import { onRequest } from 'firebase-functions/v2/https';
import express from 'express';
import { AppModule } from './app.module';

const expressApp = express();

const bootstrap = NestFactory.create(
  AppModule,
  new ExpressAdapter(expressApp),
  { logger: ['error', 'warn', 'log'] },
).then((app) => {
  app.enableShutdownHooks();
  return app.init();
});

export const api = onRequest(
  {
    region: 'asia-southeast1',
    cors: true,
    invoker: 'public',
    memory: '512MiB',
  },
  async (req, res) => {
    await bootstrap;
    expressApp(req, res);
  },
);
