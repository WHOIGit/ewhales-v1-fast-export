import sys

def convert(input_file, output_file, batch_size=2000):
    batch = []
    current_prefix = None
    
    with open(input_file, 'r', encoding='utf-8', errors='surrogateescape') as f_in, \
         open(output_file, 'w', encoding='utf-8', errors='surrogateescape') as f_out:
         
        def flush():
            if batch:
                f_out.write(current_prefix + " (" + "), (".join(batch) + ");\n")
                batch.clear()
                
        for line in f_in:
            stripped = line.strip()
            
            # Simple check for a typical mysqldump single row insert
            if stripped.startswith("INSERT INTO `") and stripped.endswith(");"):
                val_idx = stripped.find(" VALUES")
                if val_idx != -1:
                    paren_idx = stripped.find("(", val_idx)
                    if paren_idx != -1:
                        prefix = stripped[:paren_idx].strip()
                        values = stripped[paren_idx+1:-2]
                        
                        if current_prefix != prefix:
                            flush()
                            current_prefix = prefix
                            
                        batch.append(values)
                        
                        if len(batch) >= batch_size:
                            flush()
                        continue
            
            # If we fall through, flush the current batch and write the line out verbatim
            flush()
            f_out.write(line)
            
        flush()

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python convert_inserts.py <input_file.sql> <output_file.sql>")
        sys.exit(1)
    convert(sys.argv[1], sys.argv[2])
