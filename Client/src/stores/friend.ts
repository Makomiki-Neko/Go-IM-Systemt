import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { FriendInfo } from '@/types'

export const useFriendStore = defineStore('friend', () => {
  const friendList = ref<FriendInfo[]>([])
  const applyList = ref<FriendInfo[]>([])
  const pulledSet = ref<Set<string>>(new Set())

  function setFriendList(list: FriendInfo[]) { friendList.value = list }
  function setApplyList(list: FriendInfo[]) { applyList.value = list }
  function removeApply(requestId: string) {
    applyList.value = applyList.value.filter((a) => a.id !== requestId)
  }

  function addFriend(friend: FriendInfo) {
    if (!friendList.value.find((f) => f.id === friend.id)) {
      friendList.value.push(friend)
    }
  }
  function removeFriend(friendId: string) {
    friendList.value = friendList.value.filter((f) => f.id !== friendId)
  }

  function markPulled(userId: string) { pulledSet.value.add(userId) }
  function hasPulled(userId: string): boolean { return pulledSet.value.has(userId) }
  function resetPulled() { pulledSet.value = new Set() }

  return {
    friendList, applyList, pulledSet,
    setFriendList, setApplyList, removeApply,
    addFriend, removeFriend,
    markPulled, hasPulled, resetPulled,
  }
})
