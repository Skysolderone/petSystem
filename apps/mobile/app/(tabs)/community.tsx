import * as Haptics from "expo-haptics";
import { useState } from "react";
import { Text, View } from "react-native";

import { PostCard } from "@/components/community/post-card";
import { PrimaryButton } from "@/components/ui/primary-button";
import { Screen } from "@/components/ui/screen";
import { SectionCard } from "@/components/ui/section-card";
import { TextField } from "@/components/ui/text-field";
import { theme, useAppPalette } from "@/constants/theme";
import { useAskCommunityAI, useCommunityFeed, useCreatePost } from "@/services/queries/use-community";
import { usePetsList } from "@/services/queries/use-pets";
import { usePetStore } from "@/stores/pet-store";
import { ApiError } from "@/types/api";
import { useI18n } from "@/utils/i18n";

export default function CommunityScreen() {
  const palette = useAppPalette();
  const { t } = useI18n();
  const postsQuery = useCommunityFeed();
  const createPost = useCreatePost();
  const askCommunityAI = useAskCommunityAI();
  const petsQuery = usePetsList();
  const pets = petsQuery.data?.pages.flatMap((page) => page.items) ?? [];
  const selectedPetId = usePetStore((state) => state.selectedPetId);
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [tags, setTags] = useState("");
  const [question, setQuestion] = useState("");

  async function handleCreatePost() {
    await createPost.mutateAsync({
      pet_id: selectedPetId ?? pets[0]?.id,
      title,
      content,
      tags: tags ? tags.split(",").map((item) => item.trim()).filter(Boolean) : [],
    });
    await Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success).catch(() => null);
    setTitle("");
    setContent("");
    setTags("");
  }

  async function handleAskAI() {
    await askCommunityAI.mutateAsync(question);
  }

  const postError = createPost.error instanceof ApiError ? createPost.error.message : undefined;

  return (
    <Screen>
      <SectionCard title={t("community.createTitle")} subtitle={t("community.createSubtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          <TextField label={t("community.form.title")} value={title} onChangeText={setTitle} placeholder={t("community.form.titlePlaceholder")} />
          <TextField
            label={t("community.form.content")}
            value={content}
            onChangeText={setContent}
            placeholder={t("community.form.contentPlaceholder")}
          />
          <TextField label={t("community.form.tags")} value={tags} onChangeText={setTags} placeholder={t("community.form.tagsPlaceholder")} />
          {postError ? (
            <Text selectable style={{ color: palette.error }}>
              {postError}
            </Text>
          ) : null}
          <PrimaryButton label={t("community.form.submit")} onPress={handleCreatePost} loading={createPost.isPending} disabled={!content} />
        </View>
      </SectionCard>

      <SectionCard title={t("community.aiTitle")} subtitle={t("community.aiSubtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          <TextField label={t("community.aiQuestion")} value={question} onChangeText={setQuestion} placeholder={t("community.aiQuestionPlaceholder")} />
          <PrimaryButton label={t("community.aiSubmit")} onPress={handleAskAI} loading={askCommunityAI.isPending} disabled={!question} />
          {askCommunityAI.data?.answer ? (
            <Text selectable style={{ color: palette.textSecondary, lineHeight: 22 }}>
              {askCommunityAI.data.answer}
            </Text>
          ) : null}
        </View>
      </SectionCard>

      <SectionCard title={t("community.feedTitle")} subtitle={t("community.feedSubtitle")}>
        <View style={{ gap: theme.spacing.md }}>
          {(postsQuery.data ?? []).length > 0 ? (
            postsQuery.data?.map((post) => <PostCard key={post.id} post={post} />)
          ) : (
            <Text selectable style={{ color: palette.textSecondary }}>
              {t("community.feedEmpty")}
            </Text>
          )}
        </View>
      </SectionCard>
    </Screen>
  );
}
