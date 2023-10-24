import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

public class test {
    public static void main(String[] args) {
        try {
            String line;
            BufferedReader reader;
            // Go解析脚本
            System.out.println("执行检测, 线程名字为 = " + Thread.currentThread().getName());
            // 构建进程
            ProcessBuilder processBuilder = new ProcessBuilder("go", "run", "main.go", "--pcap_path", "lo", "--mode", "1", "--taskid", "0");
            processBuilder.redirectErrorStream(true); // 合并标准输出和标准错误流
            Process process = processBuilder.start();
            System.out.println("检测脚本运行的PID为:" + process.pid());

            ScheduledExecutorService executor = Executors.newScheduledThreadPool(1);
            executor.schedule(() -> {
                System.out.println("停止脚本运行的PID为:" + process.pid());
                process.destroy();
            }, 2, TimeUnit.SECONDS);
            executor.shutdown();


            reader = new BufferedReader(new InputStreamReader(process.getInputStream(), StandardCharsets.UTF_8));
            while ((line = reader.readLine()) != null) {
                System.out.println(line);
            }
            int exitCode = process.waitFor();
            System.out.println("脚本执行完毕, 退出码：" + exitCode);
            reader.close();
        } catch (IOException | InterruptedException e) {
            e.printStackTrace();
        }
    }
}
