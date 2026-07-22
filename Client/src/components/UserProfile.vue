<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { updateUserInfoApi, getUserInfoApi, updateAvatarApi } from '@/api/user'
import { getAvatarUrl } from '@/utils/avatar'

const router = useRouter()
const auth = useAuthStore()

const editing = ref(false)
const name = ref('')
const signature = ref('')
const gender = ref('')
const birthday = ref('')

const fileInput = ref<HTMLInputElement | null>(null)
const showCrop = ref(false)
const cropImageSrc = ref('')
const cropFile = ref<File | null>(null)
const cropX = ref(0)
const cropY = ref(0)
const cropSize = ref(200)
const imageNaturalW = ref(0)
const imageNaturalH = ref(0)
const imageDisplayW = ref(0)
const imageDisplayH = ref(0)

const isDragging = ref(false)
const dragStartX = ref(0)
const dragStartY = ref(0)
const cropStartX = ref(0)
const cropStartY = ref(0)

function tsToDateStr(ts: number): string {
  if (!ts) return ''
  return new Date(Number(ts) * 1000).toISOString().split('T')[0]
}

function dateStrToTs(s: string): number {
  if (!s) return 0
  return new Date(s).getTime() / 1000
}

onMounted(() => {
  if (!auth.userInfo) {
    getUserInfoApi(auth.account).then((res) => {
      if (res.data) auth.setUserInfo(res.data)
    })
  }
  name.value = auth.userInfo?.name || ''
  signature.value = auth.userInfo?.signature || ''
  gender.value = auth.userInfo?.gender || ''
  birthday.value = tsToDateStr(auth.userInfo?.birthday || 0)
})

async function save() {
  try {
    await updateUserInfoApi({
      name: name.value,
      photo: auth.userInfo?.photo || '',
      gender: gender.value,
      birthday: dateStrToTs(birthday.value),
      signature: signature.value,
    })
    auth.setUserInfo({
      ...auth.userInfo!,
      name: name.value,
      signature: signature.value,
      gender: gender.value,
      birthday: dateStrToTs(birthday.value),
    })
    editing.value = false
  } catch {}
}

function doLogout() {
  auth.logout()
  router.replace('/login')
}

/* ---- Mouse drag/zoom crop ---- */
function onFileChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  if (fileInput.value) fileInput.value.value = ''

  const reader = new FileReader()
  reader.onload = () => {
    const img = new Image()
    img.onload = () => {
      imageNaturalW.value = img.naturalWidth
      imageNaturalH.value = img.naturalHeight
      const maxDisplay = 400
      const scale = Math.min(maxDisplay / img.naturalWidth, maxDisplay / img.naturalHeight, 1)
      imageDisplayW.value = Math.round(img.naturalWidth * scale)
      imageDisplayH.value = Math.round(img.naturalHeight * scale)
      cropX.value = 0
      cropY.value = 0
      cropSize.value = 200
      cropFile.value = file
      cropImageSrc.value = reader.result as string
      showCrop.value = true
    }
    img.src = reader.result as string
  }
  reader.readAsDataURL(file)
}

function startDrag(ev: MouseEvent) {
  isDragging.value = true
  dragStartX.value = ev.clientX
  dragStartY.value = ev.clientY
  cropStartX.value = cropX.value
  cropStartY.value = cropY.value
}

function onDrag(ev: MouseEvent) {
  if (!isDragging.value) return
  const dx = ev.clientX - dragStartX.value
  const dy = ev.clientY - dragStartY.value
  const maxX = Math.max(0, imageDisplayW.value - cropSize.value)
  const maxY = Math.max(0, imageDisplayH.value - cropSize.value)
  cropX.value = Math.min(maxX, Math.max(0, cropStartX.value + dx))
  cropY.value = Math.min(maxY, Math.max(0, cropStartY.value + dy))
}

function stopDrag() {
  isDragging.value = false
}

function adjustZoom(delta: number) {
  const newSize = Math.min(imageDisplayW.value, imageDisplayH.value, Math.max(50, cropSize.value + delta))
  const maxX = Math.max(0, imageDisplayW.value - newSize)
  const maxY = Math.max(0, imageDisplayH.value - newSize)
  cropX.value = Math.min(maxX, Math.max(0, cropX.value))
  cropY.value = Math.min(maxY, Math.max(0, cropY.value))
  cropSize.value = newSize
}

async function confirmCrop() {
  if (!cropFile.value || !cropImageSrc.value) return
  const scaleX = imageNaturalW.value / imageDisplayW.value
  const scaleY = imageNaturalH.value / imageDisplayH.value

  const canvas = document.createElement('canvas')
  canvas.width = cropSize.value
  canvas.height = cropSize.value
  const ctx = canvas.getContext('2d')!

  const img = new Image()
  img.src = cropImageSrc.value
  await new Promise((resolve) => { img.onload = resolve })

  ctx.drawImage(
    img,
    cropX.value * scaleX, cropY.value * scaleY,
    cropSize.value * scaleX, cropSize.value * scaleY,
    0, 0, cropSize.value, cropSize.value,
  )

  canvas.toBlob(async (blob) => {
    if (!blob) return
    const fd = new FormData()
    fd.append('photo', blob, 'avatar.jpg')
    try {
      const { data } = await updateAvatarApi(fd)
      if (data.code === 100 && data.photo_id) {
        const res = await getUserInfoApi(auth.account)
        if (res.data) {
          auth.setUserInfo(res.data)
          auth.setPhoto(res.data.photo)
        }
      }
    } catch {}
    showCrop.value = false
  }, 'image/jpeg', 0.9)
}
</script>

<template>
  <div class="profile-panel">
    <div class="panel-header"><h3>个人信息</h3></div>
    <div class="profile-content">
      <div class="profile-avatar-large" @click="fileInput?.click()" style="cursor:pointer">
        <img v-if="getAvatarUrl(auth.userInfo?.photo)" :src="getAvatarUrl(auth.userInfo?.photo)" class="avatar-img avatar-large" />
        <span v-else>{{ auth.userInfo?.name?.[0] || auth.account?.[0] || 'U' }}</span>
        <div class="avatar-overlay">更换头像</div>
      </div>
      <input ref="fileInput" type="file" accept="image/*" style="display:none" @change="onFileChange" />

      <div class="profile-field">
        <label>账号</label>
        <span>{{ auth.account }}</span>
      </div>

      <template v-if="editing">
        <div class="profile-field">
          <label>昵称</label>
          <input v-model="name" />
        </div>
        <div class="profile-field">
          <label>签名</label>
          <input v-model="signature" />
        </div>
        <div class="profile-field">
          <label>性别</label>
          <select v-model="gender">
            <option value="">保密</option>
            <option value="male">男</option>
            <option value="female">女</option>
          </select>
        </div>
        <div class="profile-field">
          <label>生日</label>
          <input v-model="birthday" type="date" />
        </div>
        <div class="profile-actions">
          <button class="btn-primary" @click="save">保存</button>
          <button class="btn-secondary" @click="editing = false">取消</button>
        </div>
      </template>

      <template v-else>
        <div class="profile-field">
          <label>昵称</label>
          <span>{{ auth.userInfo?.name || '未设置' }}</span>
        </div>
        <div class="profile-field">
          <label>签名</label>
          <span>{{ auth.userInfo?.signature || '未设置' }}</span>
        </div>
        <div class="profile-field">
          <label>性别</label>
          <span>{{ gender === 'male' ? '男' : gender === 'female' ? '女' : '保密' }}</span>
        </div>
        <div class="profile-field">
          <label>生日</label>
          <span>{{ birthday || '未设置' }}</span>
        </div>
        <div class="profile-field">
          <label>注册时间</label>
          <span>{{ auth.userInfo?.createdtime ? new Date(auth.userInfo.createdtime * 1000).toLocaleDateString() : '-' }}</span>
        </div>
        <div class="profile-actions">
          <button class="btn-primary" @click="editing = true">编辑资料</button>
          <button class="btn-secondary" @click="doLogout">退出登录</button>
        </div>
      </template>
    </div>

    <!-- 裁剪弹窗 -->
    <div v-if="showCrop" class="crop-overlay" @mousemove="onDrag" @mouseup="stopDrag" @mouseleave="stopDrag">
      <div class="crop-dialog">
        <h3>裁剪头像</h3>
        <p class="crop-hint">拖拽图片调整裁剪区域 · 滚轮缩放</p>
        <div class="crop-container">
          <div
            class="crop-image-wrap"
            @mousedown="startDrag"
            @wheel.prevent="adjustZoom(-Math.sign(($event as WheelEvent).deltaY) * 10)"
          >
            <img
              :src="cropImageSrc"
              :style="{
                width: imageDisplayW + 'px',
                height: imageDisplayH + 'px',
                marginLeft: -cropX + 'px',
                marginTop: -cropY + 'px',
                cursor: isDragging ? 'grabbing' : 'grab',
              }"
              class="crop-image"
            />
            <div class="crop-frame" :style="{ width: cropSize + 'px', height: cropSize + 'px' }" />
          </div>
          <div class="crop-zoom-control">
            <label>缩放</label>
            <input
              type="range"
              :min="50"
              :max="Math.min(imageDisplayW, imageDisplayH)"
              :value="cropSize"
              @input="adjustZoom(Number(($event.target as HTMLInputElement).value) - cropSize)"
              class="zoom-slider"
            />
          </div>
        </div>
        <div class="crop-actions">
          <button class="btn-primary" @click="confirmCrop">确认裁剪</button>
          <button class="btn-secondary" @click="showCrop = false">取消</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.avatar-overlay {
  position: absolute; bottom: 0; left: 0; right: 0;
  background: rgba(0,0,0,0.5); color: #fff; font-size: 12px;
  text-align: center; padding: 4px 0; border-radius: 0 0 50% 50%;
  opacity: 0; transition: opacity 0.2s;
}
.profile-avatar-large { position: relative; }
.profile-avatar-large:hover .avatar-overlay { opacity: 1; }

/* 裁剪弹窗 */
.crop-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.5);
  display: flex; align-items: center; justify-content: center; z-index: 1000;
}
.crop-dialog {
  background: #fff; border-radius: 12px; padding: 24px;
  min-width: 480px; text-align: center;
}
.crop-dialog h3 { margin-bottom: 4px; }
.crop-hint { font-size: 12px; color: #999; margin-bottom: 12px; }
.crop-container {
  display: flex; flex-direction: column;
  align-items: center; gap: 12px;
}
.crop-image-wrap {
  position: relative; overflow: hidden;
  width: 400px; height: 400px; background: #f0f0f0;
  display: flex; align-items: center; justify-content: center;
  user-select: none;
}
.crop-image { max-width: none; flex-shrink: 0; pointer-events: none; }
.crop-frame {
  position: absolute; top: 50%; left: 50%;
  transform: translate(-50%, -50%);
  border: 2px solid #667eea; box-shadow: 0 0 0 9999px rgba(0,0,0,0.3);
  pointer-events: none;
}
.crop-zoom-control { display: flex; align-items: center; gap: 8px; }
.crop-zoom-control label { font-size: 13px; color: #666; }
.zoom-slider { width: 200px; }
.crop-actions { display: flex; gap: 10px; justify-content: center; margin-top: 16px; }
</style>
