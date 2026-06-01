import os
import argparse
from openai import OpenAI

def get_client(provider_name):
    """
    คืนค่า Client ของ OpenAI ที่ตั้งค่า Base URL และ API Key 
    ตาม Provider ที่เราต้องการทดสอบ
    """
    if provider_name == "gomodel":
        # ต่อผ่าน GoModel Gateway
        return OpenAI(
            base_url="http://localhost:8080/v1",
            api_key=os.getenv("GOMODEL_API_KEY", "not-needed")
        )
    elif provider_name == "openai":
        # ต่อตรงหา OpenAI
        return OpenAI(
            api_key=os.getenv("OPENAI_API_KEY", "your-openai-api-key")
        )
    elif provider_name == "groq":
        # ต่อตรงหา Groq
        return OpenAI(
            base_url="https://api.groq.com/openai/v1",
            api_key=os.getenv("GROQ_API_KEY", "your-groq-api-key")
        )
    elif provider_name == "ollama":
        # ต่อตรงหา Ollama แบบไม่ผ่าน GoModel
        return OpenAI(
            base_url="http://localhost:11434/v1",
            api_key="not-needed"
        )
    else:
        raise ValueError(f"ไม่รู้จัก Provider: {provider_name}")

def main():
    parser = argparse.ArgumentParser(description="System Test AI-LLM API")
    parser.add_argument("--provider", type=str, default="gomodel", choices=["gomodel", "openai", "groq", "ollama"], help="Provider ที่ต้องการทดสอบ")
    parser.add_argument("--model", type=str, help="ระบุชื่อ Model แบบเจาะจง (ถ้าไม่ระบุ จะดึงตัวแรกที่เจอ)")
    parser.add_argument("--prompt", type=str, default="สวัสดีครับ ขอคำแนะนำในการเขียนโปรแกรม 1 ข้อสั้นๆ", help="ข้อความทดสอบ")
    args = parser.parse_args()

    print(f"[*] Preparing to connect to Provider: {args.provider.upper()}")
    client = get_client(args.provider)
    
    try:
        # 1. ทดสอบ Endpoint: /v1/models (ดึงรายชื่อโมเดล)
        print("\n[Test 1] ดึงรายชื่อโมเดล (Models List)")
        models = client.models.list()
        model_id = args.model
        
        if models.data:
            print(f"[OK] สำเร็จ! พบโมเดลทั้งหมด {len(models.data)} ตัว")
            # แสดงแค่ 5 ตัวแรกเพื่อไม่ให้รกหน้าจอ
            for m in models.data[:5]:
                print(f"  - {m.id}")
            if len(models.data) > 5:
                print("  - ... (และอื่นๆ)")
                
            if not model_id:
                model_id = models.data[0].id
        else:
            print("[WARN] ไม่พบโมเดลในระบบ")
            if not model_id:
                print("[ERROR] กรุณาระบุชื่อโมเดลด้วย --model <model_name>")
                return

        # 2. ทดสอบ Endpoint: /v1/chat/completions (แชท)
        print(f"\n[Test 2] ส่งข้อความทดสอบ (Chat Completion) ไปยังโมเดล: {model_id}")
        response = client.chat.completions.create(
            model=model_id,
            messages=[
                {"role": "system", "content": "You are a helpful AI assistant. Answer in Thai concisely."},
                {"role": "user", "content": args.prompt}
            ],
            stream=True # เปิดโหมดพิมพ์ตอบแบบ Streaming
        )
        
        print("\n[AI Response]:")
        print("=" * 50)
        for chunk in response:
            if chunk.choices and len(chunk.choices) > 0:
                content = chunk.choices[0].delta.content
                if content:
                    print(content, end="", flush=True)
            
            # ตรวจสอบว่ามีข้อมูล Usage ส่งมาใน Chunk นี้หรือไม่
            if getattr(chunk, 'usage', None):
                usage = chunk.usage
                print(f"\n\n[📊 Token Usage]: Prompt: {usage.prompt_tokens} | Completion: {usage.completion_tokens} | Total: {usage.total_tokens}")
        print("\n" + "=" * 50)
        
    except Exception as e:
        print(f"\n[ERROR] เกิดข้อผิดพลาดในการเรียก API: {e}")

if __name__ == "__main__":
    main()
