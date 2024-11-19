<template>
  <div>
    <!-- Tag List -->
    <div class="flex flex-wrap gap-2">
      <span
        v-for="tag in modelValue"
        :key="tag"
        :class="[
          'inline-flex items-center rounded-full px-2 py-1 text-xs font-medium',
          'bg-indigo-100 text-indigo-700 dark:bg-indigo-900 dark:text-indigo-300',
        ]"
      >
        {{ tag }}
        <button
          type="button"
          @click="removeTag(tag)"
          class="ml-1 inline-flex h-4 w-4 flex-shrink-0 items-center justify-center rounded-full hover:bg-indigo-200 hover:text-indigo-900 focus:bg-indigo-500 focus:text-white focus:outline-none dark:hover:bg-indigo-800"
        >
          <span class="sr-only">Remove {{ tag }}</span>
          <XMarkIcon class="h-3 w-3" aria-hidden="true" />
        </button>
      </span>

      <!-- Add Tag Input -->
      <div
        v-if="!disabled"
        class="relative"
        @keydown.enter.prevent="addTag"
        @keydown.tab.prevent="addTag"
        @keydown.delete="handleBackspace"
      >
        <input
          type="text"
          v-model="newTag"
          :placeholder="placeholder"
          class="block w-32 rounded-full border-0 py-1 px-2 text-xs text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 dark:bg-gray-800 dark:text-white dark:ring-gray-600 dark:placeholder:text-gray-500"
          @focus="showSuggestions = true"
          @blur="handleBlur"
        />

        <!-- Tag Suggestions -->
        <div
          v-if="showSuggestions && filteredSuggestions.length > 0"
          class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none dark:bg-gray-800 sm:text-sm"
        >
          <div
            v-for="suggestion in filteredSuggestions"
            :key="suggestion"
            @mousedown.prevent="selectSuggestion(suggestion)"
            class="relative cursor-pointer select-none py-2 pl-3 pr-9 text-gray-900 hover:bg-indigo-600 hover:text-white dark:text-white dark:hover:bg-indigo-500"
          >
            {{ suggestion }}
          </div>
        </div>
      </div>
    </div>

    <!-- Error Message -->
    <p v-if="error" class="mt-2 text-sm text-red-600 dark:text-red-400">{{ error }}</p>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { XMarkIcon } from '@heroicons/vue/20/solid'

const props = defineProps({
  modelValue: {
    type: Array,
    required: true
  },
  suggestions: {
    type: Array,
    default: () => []
  },
  placeholder: {
    type: String,
    default: 'Add tag...'
  },
  disabled: {
    type: Boolean,
    default: false
  },
  maxTags: {
    type: Number,
    default: 10
  },
  error: {
    type: String,
    default: null
  }
})

const emit = defineEmits(['update:modelValue'])

const newTag = ref('')
const showSuggestions = ref(false)

const filteredSuggestions = computed(() => {
  if (!newTag.value) return props.suggestions
  const input = newTag.value.toLowerCase()
  return props.suggestions.filter(
    tag => tag.toLowerCase().includes(input) && !props.modelValue.includes(tag)
  )
})

function addTag() {
  const tag = newTag.value.trim()
  if (!tag) return
  
  if (props.modelValue.length >= props.maxTags) {
    return
  }
  
  if (!props.modelValue.includes(tag)) {
    emit('update:modelValue', [...props.modelValue, tag])
  }
  
  newTag.value = ''
  showSuggestions.value = false
}

function removeTag(tag) {
  emit('update:modelValue', props.modelValue.filter(t => t !== tag))
}

function selectSuggestion(tag) {
  if (props.modelValue.length < props.maxTags && !props.modelValue.includes(tag)) {
    emit('update:modelValue', [...props.modelValue, tag])
  }
  newTag.value = ''
  showSuggestions.value = false
}

function handleBackspace(event) {
  if (!newTag.value && props.modelValue.length > 0) {
    event.preventDefault()
    emit('update:modelValue', props.modelValue.slice(0, -1))
  }
}

function handleBlur() {
  // 延迟关闭建议列表，以便可以点击建议
  setTimeout(() => {
    showSuggestions.value = false
    // 如果有未添加的标签，添加它
    if (newTag.value) {
      addTag()
    }
  }, 200)
}
</script>
