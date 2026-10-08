<script setup lang="ts">
import { computed } from 'vue'
import { MessageCircle, Users, Sparkles, Settings2, LogOut, Search, Plus, ArrowUpRight, ArrowLeft, ArrowRight, ChevronDown, ChevronRight, X, Check, CheckCheck, Send, Paperclip, Smile, AtSign, MoreHorizontal, Bell, RefreshCw, Image, FileText, Download, LoaderCircle, AlertCircle, WifiOff, ShieldCheck, UserRound, UserPlus, Hash, Link, Copy, Sun, Moon, Mail, LockKeyhole, Eye, EyeOff, Leaf, PanelRightClose, Coffee, Compass, CircleHelp, Camera, Pencil, Trash2, CheckCircle2 } from 'lucide-vue-next'
const props = withDefaults(defineProps<{ name: string; size?: number }>(), { size: 20 })
const icons = { MessageCircle, Users, Sparkles, Settings2, LogOut, Search, Plus, ArrowUpRight, ArrowLeft, ArrowRight, ChevronDown, ChevronRight, X, Check, CheckCheck, Send, Paperclip, Smile, AtSign, MoreHorizontal, Bell, RefreshCw, Image, FileText, Download, LoaderCircle, AlertCircle, WifiOff, ShieldCheck, UserRound, UserPlus, Hash, Link, Copy, Sun, Moon, Mail, LockKeyhole, Eye, EyeOff, Leaf, PanelRightClose, Coffee, Compass, CircleHelp, Camera, Pencil, Trash2, CheckCircle2 }
const icon = computed(() => icons[props.name as keyof typeof icons] || MessageCircle)
</script>
<template><component :is="icon" :size="size" :stroke-width="1.65" aria-hidden="true" /></template>
