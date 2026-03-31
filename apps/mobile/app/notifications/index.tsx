import * as Haptics from "expo-haptics";
import { FlashList } from "@shopify/flash-list";
import { Stack } from "expo-router";
import { Text, View } from "react-native";

import { PrimaryButton } from "@/components/ui/primary-button";
import { Screen } from "@/components/ui/screen";
import { SectionCard } from "@/components/ui/section-card";
import { theme, useAppPalette } from "@/constants/theme";
import { useMarkAllNotificationsRead, useMarkNotificationRead, useNotificationsList } from "@/services/queries/use-notifications";
import { useI18n } from "@/utils/i18n";

function NotificationRow({
  id,
  title,
  body,
  createdAt,
  isRead,
}: {
  id: string;
  title: string;
  body: string;
  createdAt: string;
  isRead: boolean;
}) {
  const palette = useAppPalette();
  const { t } = useI18n();
  const markNotificationRead = useMarkNotificationRead(id);

  async function handleRead() {
    await markNotificationRead.mutateAsync();
    await Haptics.selectionAsync().catch(() => null);
  }

  return (
    <View
      style={{
        borderRadius: theme.radius.lg,
        borderCurve: "continuous",
        padding: theme.spacing.md,
        gap: theme.spacing.sm,
        backgroundColor: isRead ? palette.surface : palette.surfaceMuted,
        borderWidth: 1,
        borderColor: palette.border,
      }}
    >
      <Text selectable style={{ color: palette.text, fontWeight: "700" }}>
        {title}
      </Text>
      <Text selectable style={{ color: palette.textSecondary, lineHeight: 22 }}>
        {body}
      </Text>
      <Text selectable style={{ color: palette.textSecondary }}>
        {new Date(createdAt).toLocaleString()}
      </Text>
      {!isRead ? <PrimaryButton label={t("notifications.markRead")} onPress={handleRead} loading={markNotificationRead.isPending} /> : null}
    </View>
  );
}

export default function NotificationsScreen() {
  const palette = useAppPalette();
  const { t } = useI18n();
  const notificationsQuery = useNotificationsList();
  const markAllRead = useMarkAllNotificationsRead();
  const items = notificationsQuery.data ?? [];
  const unreadCount = items.filter((item) => !item.is_read).length;

  return (
    <Screen>
      <Stack.Screen options={{ title: t("notifications.title") }} />

      <SectionCard title={t("notifications.title")} subtitle={t("notifications.summary", { total: items.length, unread: unreadCount })}>
        <PrimaryButton
          label={t("notifications.markAll")}
          onPress={async () => {
            await markAllRead.mutateAsync();
          }}
          loading={markAllRead.isPending}
        />
      </SectionCard>

      <SectionCard title={t("notifications.listTitle")} subtitle={t("notifications.listSubtitle")}>
        {items.length > 0 ? (
          <View style={{ height: Math.max(320, items.length * 180) }}>
            <FlashList
              data={items}
              estimatedItemSize={180}
              ItemSeparatorComponent={() => <View style={{ height: theme.spacing.md }} />}
              renderItem={({ item }) => (
                <NotificationRow
                  id={item.id}
                  title={item.title}
                  body={item.body}
                  createdAt={item.created_at}
                  isRead={item.is_read}
                />
              )}
            />
          </View>
        ) : (
          <Text selectable style={{ color: palette.textSecondary }}>
            {t("notifications.empty")}
          </Text>
        )}
      </SectionCard>
    </Screen>
  );
}
