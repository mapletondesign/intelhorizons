export interface Article {
  ID: number
  URLHash: string
  URL: string
  Title: string
  Summary: string
  Author: string
  SourceID: number
  SourceName: string
  PublishedAt: string
  FetchedAt: string
  Score: number
  Category: string
}

export interface Source {
  ID: number
  Name: string
  Type: string
  URL: string
  Config: Record<string, string>
  Active: boolean
  FetchIntervalMins: number
  LastFetchedAt: string | null
}
