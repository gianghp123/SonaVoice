def get_custom_error_message(error: str, service_name: str = "Service") -> str:
    error_msg = str(error)

    if "402" in error_msg:
        return f"{service_name} unavailable: payment required. Please contact support."
    elif "429" in error_msg:
        return f"{service_name} unavailable: rate limit exceeded. Please try again later."
    elif "401" in error_msg:
        return f"{service_name} unavailable: authentication failed. Please contact support."
    elif "403" in error_msg:
        return f"{service_name} unavailable: access forbidden. Please contact support."
    else:
        return f"{service_name} unavailable: an unexpected error occurred. Please try again."