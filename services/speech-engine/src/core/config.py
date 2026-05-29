import os
from typing import Any, Dict
from pydantic_settings import BaseSettings, SettingsConfigDict
from dotenv import load_dotenv

# Load .env file if it exists
load_dotenv()

class Settings(BaseSettings):
    """
    Application settings and environment variables.
    Pydantic automatically looks for uppercase versions of these in your .env
    """
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")

    # API Keys
    OPENCODE_API_KEY: str
    OPENAI_BASE_URL: str = "https://opencode.ai/zen/go/v1"
    DEEPGRAM_API_KEY: str
    GOOGLE_API_KEY: str # Required for the embedding provider
    
    # Database
    DATABASE_URL: str
    
    LLM_NAME: str = "deepseek-v4-flash"
    
    BOT_NAME: str = "SONA"

    # Sentry
    SENTRY_DSN: str = ""
    
    # ICE / TURN Servers (JSON array of IceServer objects)
    ICE_SERVERS: str = ""

    # Constants
    EMBEDDING_DIMS: int = 768

    @property
    def memory_config(self) -> Dict[str, Any]:
        """
        Returns the Mem0 configuration dictionary.
        Using a property ensures it uses the latest settings.
        """
        return {
            "llm": {
                "provider": "openai",
                "config": {
                    "model": "deepseek-v4-flash",
                    "api_key": self.OPENCODE_API_KEY,
                    "openai_base_url": self.OPENAI_BASE_URL,
                }
            },
            "vector_store": {
                "provider": "pgvector",
                "config": {
                    "connection_string": self.DATABASE_URL,
                    "embedding_model_dims": self.EMBEDDING_DIMS,
                }
            },
            "embedder": {
                "provider": "gemini",
                "config": {
                    "model": "models/gemini-embedding-001",
                    "embedding_dims": self.EMBEDDING_DIMS,
                    "api_key": self.GOOGLE_API_KEY,
                }
            }
        }

# Instantiate settings to be used across the app
settings = Settings()