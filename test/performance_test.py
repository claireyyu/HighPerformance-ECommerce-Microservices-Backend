#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import requests
import time
import concurrent.futures
import json
import csv
import os
import matplotlib.pyplot as plt
import numpy as np
from datetime import datetime
from config import (
    ORDER_SERVICE_URL,
    THREAD_GROUPS,
    THREADS_PER_GROUP,
    REQUESTS_PER_THREAD,
    PRODUCT_ID,
    USER_ID,
    TEST_TYPES,
    RESULTS_DIR
)

# Calculate total requests
TOTAL_REQUESTS = THREAD_GROUPS * THREADS_PER_GROUP * REQUESTS_PER_THREAD

# Create results directory
os.makedirs(RESULTS_DIR, exist_ok=True)

def create_order(order_type):
    """Create a single order and return the response time and success status"""
    url = f"{ORDER_SERVICE_URL}{TEST_TYPES[order_type]}"
    payload = {
        "user_id": USER_ID,
        "product_id": PRODUCT_ID,
        "quantity": 1
    }
    
    start_time = time.time()
    try:
        response = requests.post(url, json=payload)
        end_time = time.time()
        
        if response.status_code == 200 or response.status_code == 202:
            return {
                "success": True,
                "response_time": end_time - start_time,
                "status_code": response.status_code
            }
        else:
            return {
                "success": False,
                "response_time": end_time - start_time,
                "status_code": response.status_code,
                "error": response.text
            }
    except Exception as e:
        end_time = time.time()
        return {
            "success": False,
            "response_time": end_time - start_time,
            "error": str(e)
        }

def run_test_thread(order_type):
    """Run a thread of order creation requests"""
    results = []
    for _ in range(REQUESTS_PER_THREAD):
        result = create_order(order_type)
        results.append(result)
    return results

def run_test(order_type):
    """Run a complete test for a specific order type"""
    print(f"Starting test for {order_type} order creation...")
    
    all_results = []
    start_time = time.time()
    
    # Create thread groups
    with concurrent.futures.ThreadPoolExecutor(
        max_workers=THREAD_GROUPS * THREADS_PER_GROUP
    ) as executor:
        # Submit all tasks
        futures = [
            executor.submit(run_test_thread, order_type)
            for _ in range(THREAD_GROUPS * THREADS_PER_GROUP)
        ]
        
        # Collect results as they complete
        for future in concurrent.futures.as_completed(futures):
            thread_results = future.result()
            all_results.extend(thread_results)
    
    end_time = time.time()
    total_time = end_time - start_time
    
    # Calculate metrics
    successful_requests = sum(1 for r in all_results if r["success"])
    failed_requests = len(all_results) - successful_requests
    success_rate = (successful_requests / len(all_results)) * 100
    
    response_times = [r["response_time"] for r in all_results]
    avg_response_time = sum(response_times) / len(response_times)
    median_response_time = sorted(response_times)[len(response_times) // 2]
    
    # Calculate percentiles
    sorted_times = sorted(response_times)
    p90_index = int(len(sorted_times) * 0.9)
    p95_index = int(len(sorted_times) * 0.95)
    p99_index = int(len(sorted_times) * 0.99)
    
    p90_response_time = sorted_times[p90_index]
    p95_response_time = sorted_times[p95_index]
    p99_response_time = sorted_times[p99_index]
    
    # Calculate throughput
    throughput_rps = len(all_results) / total_time
    
    # Save results to CSV
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    csv_filename = f"{RESULTS_DIR}/{order_type}_results_{timestamp}.csv"
    
    with open(csv_filename, 'w', newline='') as csvfile:
        fieldnames = [
            'request_id', 'success', 'response_time',
            'status_code', 'error'
        ]
        writer = csv.DictWriter(csvfile, fieldnames=fieldnames)
        
        writer.writeheader()
        for i, result in enumerate(all_results):
            row = {
                'request_id': i,
                'success': result['success'],
                'response_time': result['response_time'],
                'status_code': result.get('status_code', 'N/A'),
                'error': result.get('error', 'N/A')
            }
            writer.writerow(row)
    
    # Save summary metrics
    summary = {
        "order_type": order_type,
        "total_requests": len(all_results),
        "successful_requests": successful_requests,
        "failed_requests": failed_requests,
        "success_rate": success_rate,
        "total_time_seconds": total_time,
        "throughput_rps": throughput_rps,
        "avg_response_time": avg_response_time,
        "median_response_time": median_response_time,
        "p90_response_time": p90_response_time,
        "p95_response_time": p95_response_time,
        "p99_response_time": p99_response_time,
        "min_response_time": min(response_times),
        "max_response_time": max(response_times)
    }
    
    summary_filename = f"{RESULTS_DIR}/{order_type}_summary_{timestamp}.json"
    with open(summary_filename, 'w') as f:
        json.dump(summary, f, indent=4)
    
    print(f"Test completed for {order_type} order creation")
    print(f"Total requests: {len(all_results)}")
    print(f"Successful requests: {successful_requests}")
    print(f"Failed requests: {failed_requests}")
    print(f"Success rate: {success_rate:.2f}%")
    print(f"Total time: {total_time:.2f} seconds")
    print(f"Throughput: {throughput_rps:.2f} requests/second")
    print(f"Average response time: {avg_response_time:.4f} seconds")
    print(f"Median response time: {median_response_time:.4f} seconds")
    print(f"90th percentile response time: {p90_response_time:.4f} seconds")
    print(f"95th percentile response time: {p95_response_time:.4f} seconds")
    print(f"99th percentile response time: {p99_response_time:.4f} seconds")
    print(f"Min response time: {min(response_times):.4f} seconds")
    print(f"Max response time: {max(response_times):.4f} seconds")
    print(f"Results saved to {csv_filename}")
    print(f"Summary saved to {summary_filename}")
    
    return summary, all_results

def plot_results(summaries):
    """Generate plots for the test results"""
    # Create a directory for plots
    plots_dir = f"{RESULTS_DIR}/plots"
    os.makedirs(plots_dir, exist_ok=True)
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    
    # Extract data for plotting
    order_types = [s["order_type"] for s in summaries]
    throughputs = [s["throughput_rps"] for s in summaries]
    avg_response_times = [s["avg_response_time"] for s in summaries]
    p95_response_times = [s["p95_response_time"] for s in summaries]
    success_rates = [s["success_rate"] for s in summaries]
    
    # 1. Throughput comparison
    plt.figure(figsize=(10, 6))
    bars = plt.bar(order_types, throughputs, color=['blue', 'green', 'red'])
    plt.title('Throughput Comparison (Requests per Second)')
    plt.ylabel('Requests per Second')
    plt.grid(axis='y', linestyle='--', alpha=0.7)
    
    # Add value labels on top of bars
    for bar in bars:
        height = bar.get_height()
        plt.text(
            bar.get_x() + bar.get_width()/2.,
            height + 0.1,
            f'{height:.2f}',
            ha='center',
            va='bottom'
        )
    
    plt.tight_layout()
    plt.savefig(f"{plots_dir}/throughput_comparison_{timestamp}.png")
    
    # 2. Response time comparison
    plt.figure(figsize=(10, 6))
    x = np.arange(len(order_types))
    width = 0.35
    
    plt.bar(
        x - width/2,
        avg_response_times,
        width,
        label='Average',
        color='blue'
    )
    plt.bar(
        x + width/2,
        p95_response_times,
        width,
        label='95th Percentile',
        color='red'
    )
    
    plt.title('Response Time Comparison')
    plt.ylabel('Response Time (seconds)')
    plt.xticks(x, order_types)
    plt.legend()
    plt.grid(axis='y', linestyle='--', alpha=0.7)
    
    plt.tight_layout()
    plt.savefig(f"{plots_dir}/response_time_comparison_{timestamp}.png")
    
    # 3. Success rate comparison
    plt.figure(figsize=(10, 6))
    bars = plt.bar(order_types, success_rates, color=['blue', 'green', 'red'])
    plt.title('Success Rate Comparison')
    plt.ylabel('Success Rate (%)')
    plt.ylim(0, 100)
    plt.grid(axis='y', linestyle='--', alpha=0.7)
    
    # Add value labels on top of bars
    for bar in bars:
        height = bar.get_height()
        plt.text(
            bar.get_x() + bar.get_width()/2.,
            height + 1,
            f'{height:.2f}%',
            ha='center',
            va='bottom'
        )
    
    plt.tight_layout()
    plt.savefig(f"{plots_dir}/success_rate_comparison_{timestamp}.png")
    
    print(f"Plots saved to {plots_dir}")

def main():
    """Main function to run all tests"""
    print(
        f"Starting performance tests with {THREAD_GROUPS} thread groups, "
        f"{THREADS_PER_GROUP} threads per group, "
        f"{REQUESTS_PER_THREAD} requests per thread"
    )
    print(f"Total requests per test: {TOTAL_REQUESTS}")
    
    summaries = []
    all_results = {}
    
    # Run tests for each order type
    for order_type in TEST_TYPES:
        summary, results = run_test(order_type)
        summaries.append(summary)
        all_results[order_type] = results
    
    # Generate plots
    plot_results(summaries)
    
    print("All tests completed!")

if __name__ == "__main__":
    main() 