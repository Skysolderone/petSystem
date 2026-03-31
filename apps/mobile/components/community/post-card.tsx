import * as Haptics from "expo-haptics";
import { useState } from "react";
import { Text, View } from "react-native";

import { PrimaryButton } from "@/components/ui/primary-button";
import { TextField } from "@/components/ui/text-field";
import { theme, useAppPalette } from "@/constants/theme";
import { useCreateComment, usePostComments, useTogglePostLike } from "@/services/queries/use-community";
import type { Post } from "@/types/community";
import { useI18n } from "@/utils/i18n";

interface PostCardProps {
  post: Post;
}

export function PostCard({ post }: PostCardProps) {
  const palette = useAppPalette();
  const { t } = useI18n();
  const commentsQuery = usePostComments(post.id);
  const createComment = useCreateComment(post.id);
  const toggleLike = useTogglePostLike(post.id);
  const [commentDraft, setCommentDraft] = useState("");

  async function handleLike() {
    await toggleLike.mutateAsync();
    await Haptics.selectionAsync().catch(() => null);
  }

  async function handleComment() {
    await createComment.mutateAsync(commentDraft);
    setCommentDraft("");
  }

  return (
    <View
      style={{
        borderRadius: theme.radius.lg,
        borderCurve: "continuous",
        padding: theme.spacing.md,
        gap: theme.spacing.md,
        borderWidth: 1,
        borderColor: palette.border,
        backgroundColor: palette.surface,
      }}
    >
      <View style={{ gap: theme.spacing.xs }}>
        <Text selectable style={{ color: palette.text, fontWeight: "700", fontSize: theme.fontSize.md }}>
          {post.title || t("community.post.defaultTitle")}
        </Text>
        <Text selectable style={{ color: palette.textSecondary }}>
          {t("community.post.meta", {
            time: new Date(post.created_at).toLocaleString(),
            likes: post.like_count,
            comments: post.comment_count,
          })}
        </Text>
      </View>

      <Text selectable style={{ color: palette.textSecondary, lineHeight: 22 }}>
        {post.content}
      </Text>

      {post.tags.length > 0 ? (
        <Text selectable style={{ color: palette.primary }}>
          #{post.tags.join(" #")}
        </Text>
      ) : null}

      <View style={{ gap: theme.spacing.sm }}>
        <PrimaryButton label={t("community.post.like")} onPress={handleLike} variant="secondary" />
        <TextField
          label={t("community.post.commentLabel")}
          value={commentDraft}
          onChangeText={setCommentDraft}
          placeholder={t("community.post.commentPlaceholder")}
        />
        <PrimaryButton label={t("community.post.commentSubmit")} onPress={handleComment} disabled={!commentDraft} />
      </View>

      <View style={{ gap: theme.spacing.sm }}>
        {(commentsQuery.data ?? []).slice(0, 3).map((comment) => (
          <View
            key={comment.id}
            style={{
              borderRadius: theme.radius.md,
              borderCurve: "continuous",
              backgroundColor: palette.surfaceMuted,
              padding: theme.spacing.md,
              gap: theme.spacing.xs,
            }}
          >
            <Text selectable style={{ color: palette.textSecondary }}>
              {comment.content}
            </Text>
            <Text selectable style={{ color: palette.textSecondary }}>
              {new Date(comment.created_at).toLocaleString()}
            </Text>
          </View>
        ))}
      </View>
    </View>
  );
}
