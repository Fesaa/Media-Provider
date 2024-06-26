import React from "react";
import SuccesNotification from "./succes";
import ErrorNotification from "./error";

class NotificationHandler extends React.Component {
  static notifications: React.Component[] = [];
  static fadeInTime: number = 2000;
  static instance: NotificationHandler | null = null;

  constructor(props: any) {
    super(props);
    const copy = NotificationHandler.notifications;
    NotificationHandler.instance = this;
    NotificationHandler.notifications = copy;
  }

  private static forceUpdate() {
    if (NotificationHandler.instance) {
      NotificationHandler.instance.forceUpdate();
    }
  }

  public static addSuccesNotificationByTitle(title: string) {
    NotificationHandler.addNotification(
      new SuccesNotification({
        title: title,
        description: null,
      }),
    );
  }

  public static addErrorNotificationByTitle(title: string) {
    NotificationHandler.addNotification(
      new ErrorNotification({
        title: title,
        description: null,
      }),
    );
  }

  public static addNotification(notification: React.Component): void {
    NotificationHandler.notifications.push(notification);
    NotificationHandler.forceUpdate();

    setTimeout(() => {
      NotificationHandler.notifications.shift();
      NotificationHandler.forceUpdate();
    }, NotificationHandler.fadeInTime);
  }

  render(): React.ReactNode {
    return (
      <div className="fixed top-4 md:top-10 right-4 flex flex-col space-y-4">
        {NotificationHandler.notifications.map((comp, i) => (
          <div key={i}>{comp.render()}</div>
        ))}
      </div>
    );
  }
}

export default NotificationHandler;
