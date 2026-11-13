"""One LangChain structured model path for Gemini, OpenAI and Anthropic."""

from inspect import isawaitable

from langchain.chat_models import init_chat_model
from langchain_core.messages import HumanMessage, SystemMessage
from langsmith import tracing_context

from hexta_ai.providers.base import INSTRUCTIONS, ProviderError, StructuredProvider, Usage

MODEL_PROVIDERS = {"gemini": "google_genai", "openai": "openai", "anthropic": "anthropic"}


class LangChainProvider(StructuredProvider):
    def __init__(self, name, config, chat_model=None):
        super().__init__(config)
        self.name = name
        if chat_model is None:
            options = {
                "api_key": config.api_key,
                "max_tokens": config.max_output_tokens,
                "timeout": config.timeout,
                "max_retries": 0,
            }
            if name == "gemini":
                # This integration maps max_retries to Google's TOTAL attempts.
                # Zero selects the SDK default; one means no repeat request.
                options.update(max_retries=1, vertexai=False, temperature=config.temperature)
            elif name == "openai":
                options.update(use_responses_api=True, store=False, temperature=config.temperature)
            self.chat_model = init_chat_model(config.model, model_provider=MODEL_PROVIDERS[name], **options)
        else:
            self.chat_model = chat_model

    async def _request(self, schema, payload):
        structured = self.chat_model.with_structured_output(schema, method="json_schema", include_raw=True)
        with tracing_context(enabled=False):
            response = await structured.ainvoke(
                [SystemMessage(content=getattr(schema, "instructions", INSTRUCTIONS)), HumanMessage(content=payload)],
                config={"callbacks": []},
            )
        if response.get("parsing_error") is not None or response.get("parsed") is None:
            raise ProviderError("LLM_INVALID_OUTPUT")
        raw = response["raw"]
        metadata = raw.response_metadata
        finish = metadata.get("finish_reason") or metadata.get("stop_reason") or metadata.get("status")
        if finish is not None and finish not in ("STOP", "stop", "end_turn", "completed"):
            raise ProviderError("LLM_INVALID_OUTPUT")
        usage = raw.usage_metadata
        return response["parsed"], (
            Usage(input_tokens=usage["input_tokens"], output_tokens=usage["output_tokens"])
            if usage else Usage(estimated=True)
        )

    async def close(self):
        # LangChain integrations have no common close protocol. Close only clients
        # already created by this model; never instantiate a lazy client to close it.
        closer = getattr(self.chat_model, "aclose", None)
        if closer:
            await closer()
            return
        for name in ("root_async_client", "root_client", "_async_client", "_client"):
            client = vars(self.chat_model).get(name)
            if client is not None:
                result = client.close()
                if isawaitable(result):
                    await result
