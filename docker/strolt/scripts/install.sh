ls /usr/bin > initial_usr_bin_list
ls /usr/local/bin > initial_usr_local_bin_list
# /////////////////////////////////////


sh ./scripts/tools/postgresql.sh
sh ./scripts/tools/restic.sh
sh ./scripts/tools/mongodb.sh
sh ./scripts/tools/mysql.sh
# sh ./scripts/tools/borgbackup.sh




# rm /usr/bin/gdbm_dump
# rm /usr/bin/gdbm_load
# rm /usr/bin/gdbmtool

# /////////////////////////////////////

ls /usr/bin > final_usr_bin_list
ls /usr/local/bin > final_usr_local_bin_list

printf "\n\n\n\n\nNEW files in /usr/bin >>>"
diff -u initial_usr_bin_list final_usr_bin_list | grep -E "^\+"
printf "<<<\n\n"

printf "NEW files in /usr/local/bin >>>"
diff -u initial_usr_local_bin_list final_usr_local_bin_list | grep -E "^\+"
printf "<<<\n\n\n\n\n"

rm initial_usr_bin_list && rm final_usr_bin_list && rm initial_usr_local_bin_list && rm final_usr_local_bin_list
