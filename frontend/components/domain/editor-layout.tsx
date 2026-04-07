'use client'

import { useState, useTransition } from 'react'
import { AgentWorkspace, type UploadedImage } from '@/components/domain/agent-workspace'
import { XPostCard } from '@/components/domain/x-post-card'
import {
  uploadMediaAction,
  deleteMediaAction,
  type ChatMessageEntry,
  type AgentRun,
  type MediaEntry,
} from '@/app/dashboard/events/[id]/post/actions'

export function EditorLayout({
  eventId,
  initialEntries,
  initialRun,
  postContent,
  initialMedia,
}: {
  eventId: string
  initialEntries: ChatMessageEntry[]
  initialRun: AgentRun | null
  postContent: string
  initialMedia: MediaEntry[]
}) {
  const [images, setImages] = useState<UploadedImage[]>(() =>
    initialMedia.map((m) => ({ id: m.id, file: null, url: m.url }))
  )
  const [, startTransition] = useTransition()

  function handleImagesChange(nextImages: UploadedImage[]) {
    // Find newly added images (no matching id in current list)
    const currentIds = new Set(images.map((img) => img.id))
    const added = nextImages.filter((img) => !currentIds.has(img.id))

    // Find removed images
    const nextIds = new Set(nextImages.map((img) => img.id))
    const removed = images.filter((img) => !nextIds.has(img.id))

    setImages(nextImages)

    // Upload new images
    for (const img of added) {
      if (!img.file || !(img.file instanceof File)) continue
      const formData = new FormData()
      formData.append('file', img.file)
      startTransition(async () => {
        try {
          const saved = await uploadMediaAction(eventId, formData)
          // Replace only the id for delete tracking; keep local object URL for display
          setImages((prev) =>
            prev.map((i) => (i.id === img.id ? { ...i, id: saved.id } : i))
          )
        } catch (err) {
          console.error('Failed to upload image:', err)
          // Remove the failed image from state
          setImages((prev) => prev.filter((i) => i.id !== img.id))
        }
      })
    }

    // Delete removed images (only if they have a persisted id, not local-*)
    for (const img of removed) {
      if (img.id.startsWith('img-')) continue // local-only, never persisted
      startTransition(async () => {
        try {
          await deleteMediaAction(img.id)
        } catch (err) {
          console.error('Failed to delete image:', err)
        }
      })
    }
  }

  return (
    <div className="grid flex-1 min-h-0 gap-6 xl:grid-cols-2">
      <div className="min-h-0">
        <AgentWorkspace
          eventId={eventId}
          initialEntries={initialEntries}
          initialRun={initialRun}
          images={images}
          onImagesChange={handleImagesChange}
        />
      </div>

      <div className="min-h-0">
        <div className="h-full min-h-0 overflow-hidden">
          <XPostCard
            content={postContent}
            imageUrls={images.map((img) => img.url)}
          />
        </div>
      </div>
    </div>
  )
}
