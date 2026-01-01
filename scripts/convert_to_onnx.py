import argparse
from optimum.onnxruntime import ORTModelForSequenceClassification
from transformers import AutoTokenizer
import os

def convert_model(model_name, output_dir):
    """
    Downloads a model and tokenizer from Hugging Face and saves it in ONNX format.
    """
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)
        print(f"Created output directory: {output_dir}")

    # Load the tokenizer and model from the Hub and export to ONNX
    print(f"Loading model '{model_name}'...")
    model = ORTModelForSequenceClassification.from_pretrained(model_name, from_transformers=True)
    tokenizer = AutoTokenizer.from_pretrained(model_name)

    # Save the converted model and tokenizer
    print(f"Saving ONNX model and tokenizer to '{output_dir}'...")
    model.save_pretrained(output_dir)
    tokenizer.save_pretrained(output_dir)

    print("\nConversion complete!")
    print(f"Model files are located in: {output_dir}")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Convert a Hugging Face model to ONNX format.")
    parser.add_argument(
        "--model",
        type=str,
        default="distilbert-base-uncased-finetuned-sst-2-english",
        help="The name of the Hugging Face model to convert."
    )
    parser.add_argument(
        "--output",
        type=str,
        default="./models/distilbert-onnx",
        help="The directory to save the ONNX model files."
    )
    args = parser.parse_args()

    convert_model(args.model, args.output)
