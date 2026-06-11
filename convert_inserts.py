# convert_inserts.py
# 
# Intended to turn a sql dump file of single row inserts into a file of
# multiple row insert statements of up to the designated batch size.
# 
# MySQL has no constraint on how many rows can be inserted with a single
# insert statement, but it must fit within the configured max_allowed_packet
# of the server. Batch sizes need to be chosen such that the batched insert
# statements will fit within. Adjust as needed.

import sys

def convert(input_file, output_file, batch_size=2000):
    batch = []
    current_prefix = None
    
    # need to use surrogateescape to handle raw bytes in the dump
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
    if len(sys.argv) > 2:
        infile, outfile = sys.argv[1], sys.argv[2]
        batch_size = int(sys.argv[3]) if len(sys.argv) > 3 else 2000
        convert(sys.argv[1], sys.argv[2], batch_size)
        exit(0)
    print("Usage: python convert_inserts.py <input_file.sql> <output_file.sql> [<batch_size>]")
