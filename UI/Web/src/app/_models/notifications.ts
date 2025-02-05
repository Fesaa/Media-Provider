export interface Notification {
  id: number;
  title: string;
  summary: string;
  body: string;
  colour: NotificationColour;
  group: NotificationGroup;
  read: boolean;
  readAt?: Date;
}

export enum NotificationColour {
  Primary = "primary",
  Secondary = "secondary",
  Success = "success",
  Info = "info",
  Warn = "warn",
  Help = "help",
  Danger = "danger",
  Contrast = "contrast"
}

export enum NotificationGroup {
  Content = "content",
  Security = "security",
  General = "general"
}
