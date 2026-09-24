<template>
  <div class="page" v-if="note">
    <el-page-header @back="$router.back()" content="品鉴详情" />
    <el-row :gutter="16">
      <el-col :xs="24" :md="14">
        <el-card>
          <el-image v-if="note.image_url" :src="note.image_url" fit="cover" class="cover" />
          <h1>{{ displayName(note) }}</h1>
          <div class="meta">{{ displayOrigin(note) || '-' }} · {{ RoastLevelMap[note.roast_level] }} · {{ note.brew_method || '-' }}</div>
          <el-alert
            v-if="note.coffee_bean"
            class="bean-link"
            type="success"
            :closable="false"
            show-icon
            title="已绑定豆种档案"
            :description="`最新名称：${note.coffee_bean.name} · 产地：${note.coffee_bean.origin || '-'} · 处理法：${ProcessMethodMap[note.coffee_bean.process_method]}`"
          />
          <ScoreStars :model-value="note.overall_score" />
          <FlavorTags :tags="note.flavor_tags" />
          <el-descriptions :column="2" border class="scores">
            <el-descriptions-item label="香气">{{ note.aroma_score }}</el-descriptions-item>
            <el-descriptions-item label="酸质">{{ note.acidity_score }}</el-descriptions-item>
            <el-descriptions-item label="醇厚">{{ note.body_score }}</el-descriptions-item>
            <el-descriptions-item label="综合">{{ note.overall_score }}</el-descriptions-item>
          </el-descriptions>
          <p class="notes">{{ note.notes_text }}</p>
          <div class="actions">
            <el-button :type="liked ? 'warning' : 'default'" :loading="liking" @click="toggleLike">
              👍 {{ likeCount }}
            </el-button>
            <el-button v-if="isOwner" type="primary" plain @click="openEdit">编辑</el-button>
            <el-button v-if="isOwner" type="danger" plain @click="remove">删除</el-button>
          </div>
        </el-card>
        <el-card v-if="recipe" class="block">
          <template #header>关联配方：{{ recipe.name }}</template>
          <p>{{ recipe.device }} · {{ recipe.water_temp }}°C · {{ recipe.grind_size }} · 粉水比 {{ recipe.ratio }}</p>
          <ol>
            <li v-for="s in steps" :key="s.step_number">
              第{{ s.step_number }}步：{{ s.description }}（{{ s.duration_seconds }}s）
            </li>
          </ol>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="10">
        <el-card>
          <template #header>评论（{{ comments.length }}）</template>
          <div v-for="c in comments" :key="c.id" class="comment">
            <div class="c-head">用户 #{{ c.user_id }} · {{ formatDateTime(c.created_at) }}</div>
            <div>{{ c.content }}</div>
          </div>
          <el-empty v-if="!comments.length" description="暂无评论" />
          <div class="reply">
            <el-input v-model="reply" type="textarea" :rows="3" placeholder="写下你的评论…" />
            <el-button type="primary" :loading="replying" @click="submitReply">发表评论</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="showEdit" title="编辑品鉴笔记" width="560px">
      <el-form label-width="90px">
        <el-form-item label="豆种">
          <el-select v-model="editForm.beanId" placeholder="选择豆种绑定档案" filterable clearable style="width: 100%" @change="onEditBeanChange">
            <el-option v-for="b in beans" :key="b.id" :label="`${b.name}（${b.origin}）`" :value="b.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="咖啡名称"><el-input v-model="editForm.coffee_name" /></el-form-item>
        <el-form-item label="产地"><el-input v-model="editForm.origin" /></el-form-item>
        <el-form-item label="烘焙度">
          <el-radio-group v-model="editForm.roast_level">
            <el-radio-button v-for="(label, value) in RoastLevelMap" :key="value" :value="value">{{ label }}</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="香气分"><el-rate v-model="editForm.aroma_score" :max="10" show-score /></el-form-item>
        <el-form-item label="酸质分"><el-rate v-model="editForm.acidity_score" :max="10" show-score /></el-form-item>
        <el-form-item label="醇厚度"><el-rate v-model="editForm.body_score" :max="10" show-score /></el-form-item>
        <el-form-item label="综合分"><el-rate v-model="editForm.overall_score" :max="10" show-score /></el-form-item>
        <el-form-item label="冲煮方式"><el-input v-model="editForm.brew_method" /></el-form-item>
        <el-form-item label="品鉴笔记"><el-input v-model="editForm.notes_text" type="textarea" :rows="4" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEdit = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import ScoreStars from '@/components/common/ScoreStars.vue'
import FlavorTags from '@/components/common/FlavorTags.vue'
import { getNote, listComments, createComment, likeNote, unlikeNote, deleteNote, updateNote } from '@/api/note'
import { listBeans } from '@/api/bean'
import { getRecipe } from '@/api/recipe'
import { useAuth } from '@/hooks/useAuth'
import { RoastLevelMap, displayName, displayOrigin, type TastingNote } from '@/constants/note'
import { ProcessMethodMap, type CoffeeBean } from '@/constants/bean'
import type { Comment, BrewRecipe, RecipeStep } from '@/types/api'
import { formatDateTime } from '@/utils/dateFormat'

const route = useRoute()
const router = useRouter()
const { isLoggedIn, user } = useAuth()
const note = ref<TastingNote | null>(null)
const likeCount = ref(0)
const liked = ref(false)
const liking = ref(false)
const comments = ref<Comment[]>([])
const reply = ref('')
const replying = ref(false)
const recipe = ref<BrewRecipe | null>(null)
const beans = ref<CoffeeBean[]>([])
const showEdit = ref(false)
const saving = ref(false)
const editForm = reactive({
  beanId: undefined as number | undefined,
  coffee_name: '', origin: '', roast_level: 'light' as string,
  aroma_score: 0, acidity_score: 0, body_score: 0, overall_score: 0,
  brew_method: '', notes_text: '',
})

const steps = computed<RecipeStep[]>(() => {
  try {
    return JSON.parse(recipe.value?.steps || '[]')
  } catch {
    return []
  }
})
const isOwner = computed(() => !!user.value && note.value?.user_id === user.value.id)

onMounted(async () => {
  const id = route.params.id as string
  const res = await getNote(id)
  note.value = res.note
  likeCount.value = res.like_count
  comments.value = await listComments(res.note.id)
  beans.value = (await listBeans({ page_size: 100 })).list
  if (res.note.brew_recipe_id) {
    try {
      recipe.value = await getRecipe(res.note.brew_recipe_id)
    } catch {
      recipe.value = null
    }
  }
})

async function loadNote(id: string) {
  const res = await getNote(id)
  note.value = res.note
  likeCount.value = res.like_count
  if (res.note.brew_recipe_id) {
    try {
      recipe.value = await getRecipe(res.note.brew_recipe_id)
    } catch {
      recipe.value = null
    }
  }
}

function openEdit() {
  if (!note.value) return
  editForm.beanId = note.value.coffee_bean_id ?? undefined
  editForm.coffee_name = note.value.coffee_name
  editForm.origin = note.value.origin
  editForm.roast_level = note.value.roast_level
  editForm.aroma_score = note.value.aroma_score
  editForm.acidity_score = note.value.acidity_score
  editForm.body_score = note.value.body_score
  editForm.overall_score = note.value.overall_score
  editForm.brew_method = note.value.brew_method
  editForm.notes_text = note.value.notes_text
  showEdit.value = true
}

function onEditBeanChange(id: number | undefined) {
  const b = beans.value.find((x) => x.id === id)
  if (b) {
    editForm.coffee_name = b.name
    editForm.origin = b.origin
  }
}

async function saveEdit() {
  if (!editForm.coffee_name || !editForm.roast_level) {
    ElMessage.warning('请填写咖啡名称与烘焙度')
    return
  }
  saving.value = true
  try {
    // coffee_bean_id: 0 explicitly unbinds; a selected id (re)binds the note.
    await updateNote(note.value!.id, {
      coffee_bean_id: editForm.beanId ?? 0,
      coffee_name: editForm.coffee_name,
      origin: editForm.origin,
      roast_level: editForm.roast_level as TastingNote['roast_level'],
      aroma_score: editForm.aroma_score,
      acidity_score: editForm.acidity_score,
      body_score: editForm.body_score,
      overall_score: editForm.overall_score,
      brew_method: editForm.brew_method,
      notes_text: editForm.notes_text,
    })
    ElMessage.success('品鉴笔记已更新')
    showEdit.value = false
    await loadNote(String(note.value!.id))
  } finally {
    saving.value = false
  }
}

async function toggleLike() {
  if (!isLoggedIn.value) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }
  liking.value = true
  try {
    if (liked.value) {
      await unlikeNote(note.value!.id)
      liked.value = false
      likeCount.value = Math.max(0, likeCount.value - 1)
    } else {
      await likeNote(note.value!.id)
      liked.value = true
      likeCount.value += 1
    }
  } finally {
    liking.value = false
  }
}

async function submitReply() {
  if (!isLoggedIn.value) {
    ElMessage.warning('请先登录')
    router.push('/login')
    return
  }
  if (!reply.value.trim()) return
  replying.value = true
  try {
    await createComment(note.value!.id, reply.value)
    comments.value = await listComments(note.value!.id)
    reply.value = ''
  } finally {
    replying.value = false
  }
}

async function remove() {
  await deleteNote(note.value!.id)
  ElMessage.success('笔记已删除')
  router.push('/')
}
</script>

<style scoped>
.page { max-width: 1000px; margin: 0 auto; }
.cover { width: 100%; max-height: 360px; border-radius: 8px; }
.meta { color: #999; margin: 8px 0; }
.bean-link { margin: 10px 0; }
.scores { margin-top: 12px; }
.notes { line-height: 1.8; margin-top: 12px; }
.actions { margin-top: 16px; display: flex; gap: 12px; }
.block { margin-top: 16px; }
.comment { border-bottom: 1px solid #f0f0f0; padding: 10px 0; }
.c-head { color: #999; font-size: 12px; }
.reply { margin-top: 12px; display: flex; flex-direction: column; gap: 10px; }
</style>
