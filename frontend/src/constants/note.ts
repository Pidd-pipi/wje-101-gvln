import type { CoffeeBean } from './bean'

export type RoastLevel = 'light' | 'medium' | 'dark'

export const RoastLevelMap: Record<RoastLevel, string> = {
  light: '浅烘',
  medium: '中烘',
  dark: '深烘',
}

export const ROAST_LEVELS = Object.keys(RoastLevelMap) as RoastLevel[]

export interface TastingNote {
  id: number
  user_id: number
  coffee_bean_id: number | null
  coffee_bean?: CoffeeBean | null
  coffee_name: string
  origin: string
  roast_level: RoastLevel
  flavor_tags: string
  aroma_score: number
  acidity_score: number
  body_score: number
  overall_score: number
  brew_method: string
  brew_recipe_id: number
  notes_text: string
  image_url: string
  created_at: string
  updated_at: string
}

export interface NoteItem {
  note: TastingNote
  like_count: number
}

// displayName returns the latest bean name when bound, otherwise the
// snapshot name stored on the note.
export function displayName(note: TastingNote): string {
  return note.coffee_bean?.name || note.coffee_name
}

// displayOrigin returns the latest bean origin when bound, otherwise the
// snapshot origin stored on the note.
export function displayOrigin(note: TastingNote): string {
  return note.coffee_bean?.origin || note.origin
}

export function parseTags(raw: string): string[] {
  try {
    const arr = JSON.parse(raw || '[]')
    return Array.isArray(arr) ? arr : []
  } catch {
    return []
  }
}
